package relay

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hookie-sh/cli/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn      *grpc.ClientConn
	client    proto.RelayServiceClient
	token     string
	channelID string // anonymous channel ID
	anonymous bool   // anonymous mode flag
}

func NewClient(token string, debug bool) (*Client, error) {
	relayURL := os.Getenv("HOOKIE_RELAY_URL")
	if relayURL == "" {
		relayURL = GetRelayURL()
	}
	relayURL = normalizeRelayURL(relayURL)

	// Transport: TLS on port 443 (Railway custom domain); plaintext on other ports.
	creds := transportCreds(relayURL)

	conn, err := grpc.NewClient(relayURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %w", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewRelayServiceClient(conn),
		token:  token,
	}, nil
}

// transportCreds returns TLS for port 443 or for the production relay host (TCP Proxy uses a non-443 port).
// Set HOOKIE_INSECURE_TLS=1 to force plaintext for any URL.
func transportCreds(relayURL string) credentials.TransportCredentials {
	if os.Getenv("HOOKIE_INSECURE_TLS") != "" {
		return insecure.NewCredentials()
	}
	host, port := hostFromURL(relayURL)
	if port == "443" {
		return credentials.NewTLS(nil)
	}
	// Dev builds default to localhost:50051; host == defaultHost would wrongly enable TLS on plaintext relay.
	if isLocalhost(relayURL) {
		return insecure.NewCredentials()
	}
	// TCP Proxy + relay-proxy: public URL is relay.hookie.sh:<proxy_port>; still use TLS
	defaultHost, _ := hostFromURL(GetRelayURL())
	if host != "" && host == defaultHost {
		return credentials.NewTLS(nil)
	}
	return insecure.NewCredentials()
}

// isLocalhost checks if the URL is pointing to localhost or 127.0.0.1
func isLocalhost(url string) bool {
	host, _ := hostFromURL(url)
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || host == ""
}

// hostFromURL returns the host part of the URL (scheme and port stripped).
func hostFromURL(url string) (host string, port string) {
	s := strings.TrimPrefix(url, "grpc://")
	s = strings.TrimPrefix(s, "grpcs://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	var err error
	host, port, err = net.SplitHostPort(s)
	if err != nil {
		return s, ""
	}
	return host, port
}

// normalizeRelayURL ensures a non-localhost URL has port 443 for TLS (avoids defaulting to 80).
func normalizeRelayURL(url string) string {
	host, port := hostFromURL(url)
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "" {
		return url
	}
	if port != "" {
		return url
	}
	return net.JoinHostPort(host, "443")
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) createContext(ctx context.Context) context.Context {
	if c.anonymous {
		md := metadata.New(map[string]string{
			"x-channel-type": "anonymous",
		})
		return metadata.NewOutgoingContext(ctx, md)
	}
	md := metadata.New(map[string]string{
		"authorization": c.token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) ListApplications(ctx context.Context, orgID string) ([]*proto.Application, error) {
	req := &proto.ListApplicationsRequest{
		OrgId: orgID,
	}
	resp, err := c.client.ListApplications(c.createContext(ctx), req)
	if err != nil {
		return nil, err
	}
	return resp.Applications, nil
}

func (c *Client) ListSources(ctx context.Context, appID string) ([]*proto.Source, error) {
	req := &proto.ListSourcesRequest{
		AppId: appID,
	}
	resp, err := c.client.ListSources(c.createContext(ctx), req)
	if err != nil {
		return nil, err
	}
	return resp.Sources, nil
}

func (c *Client) Subscribe(ctx context.Context, appID, sourceID, orgID, machineID string) (grpc.BidiStreamingClient[proto.SubscribeMessage, proto.Event], error) {
	stream, err := c.client.Subscribe(c.createContext(ctx))
	if err != nil {
		return nil, err
	}

	// Send initial SubscribeRequest
	req := &proto.SubscribeMessage{
		Message: &proto.SubscribeMessage_Subscribe{
			Subscribe: &proto.SubscribeRequest{
				AppId:     appID,
				SourceId:  sourceID,
				OrgId:     orgID,
				MachineId: machineID,
			},
		},
	}
	if err := stream.Send(req); err != nil {
		return nil, fmt.Errorf("failed to send SubscribeRequest: %w", err)
	}

	return stream, nil
}

// NewAnonymousClient creates a new relay client for anonymous usage (no auth)
func NewAnonymousClient(debug bool) (*Client, error) {
	relayURL := os.Getenv("HOOKIE_RELAY_URL")
	if relayURL == "" {
		relayURL = GetRelayURL()
	}
	relayURL = normalizeRelayURL(relayURL)

	creds := transportCreds(relayURL)
	conn, err := grpc.NewClient(relayURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %w", err)
	}

	return &Client{
		conn:      conn,
		client:    proto.NewRelayServiceClient(conn),
		anonymous: true,
	}, nil
}

// CreateAnonymousChannel creates an anonymous ephemeral channel
func (c *Client) CreateAnonymousChannel(ctx context.Context) (*proto.CreateAnonymousChannelResponse, error) {
	return c.client.CreateAnonymousChannel(c.createContext(ctx), &proto.CreateAnonymousChannelRequest{})
}

// Ping checks connectivity to the relay (no auth required)
func (c *Client) Ping(ctx context.Context) (*proto.PingResponse, error) {
	return c.client.Ping(ctx, &proto.PingRequest{})
}

// ErrRelayHTTP2Hint is appended when the relay error looks like HTTP/1.1 in front (400/404, unexpected content-type).
const ErrRelayHTTP2Hint = " (relay requires HTTP/2: use TCP Proxy + relay-proxy and set HOOKIE_RELAY_URL=relay.hookie.sh:<TCP proxy port> — see backend/relay/README-TLS.md)"

// WrapRelayErr adds a hint to relay errors that indicate HTTP/2 was not used (400, 404, unexpected content-type).
func WrapRelayErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "400") || strings.Contains(msg, "404") || strings.Contains(msg, "unexpected content-type") {
		return fmt.Errorf("%w%s", err, ErrRelayHTTP2Hint)
	}
	return err
}

// SetChannelID sets the anonymous channel ID
func (c *Client) SetChannelID(channelID string) {
	c.channelID = channelID
}

package publicid

import "regexp"

var (
	AppPublicIDPattern  = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,28}[a-z0-9])?-[a-z0-9]{6}$`)
	SourceSlugPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,30}[a-z0-9])?$`)
	AnonPublicIDPattern = regexp.MustCompile(`^[a-z]+-[a-z]+-[a-z0-9]{6}$`)
)

func ValidAppPublicID(id string) bool {
	return AppPublicIDPattern.MatchString(id)
}

func ValidSourceSlug(slug string) bool {
	return SourceSlugPattern.MatchString(slug)
}

func ValidAnonPublicID(id string) bool {
	return AnonPublicIDPattern.MatchString(id)
}

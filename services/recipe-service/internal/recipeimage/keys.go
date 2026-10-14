package recipeimage

import "github.com/google/uuid"

// objectKeys builds the object keys for one image. Every input is a UUID, so
// user input can never reach a key.
func objectKeys(tenantID, householdID, recipeID, imageID uuid.UUID) (full, thumb string) {
	prefix := tenantID.String() + "/" + householdID.String() + "/recipes/" + recipeID.String() + "/" + imageID.String() + "/"
	return prefix + "full.jpg", prefix + "thumb.jpg"
}

// URLs returns the public, versioned image URLs. The /api/v1 prefix is the
// public path under both compose nginx and the k8s ingress. v changes on
// every replace, which busts browser caches of the immutable responses.
func URLs(recipeID, imageID uuid.UUID) (thumb, full string) {
	base := "/api/v1/recipes/" + recipeID.String() + "/image?size="
	v := "&v=" + imageID.String()
	return base + string(SizeThumb) + v, base + string(SizeFull) + v
}

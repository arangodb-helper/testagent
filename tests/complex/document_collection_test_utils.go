package complex

import "strconv"

// Equals returns true when the value fields of `d` and `other` are the equal.
func (d BigDocument) Equals(other BigDocument) bool {
	return d.Value == other.Value &&
		d.Name == other.Name &&
		d.Odd == other.Odd &&
		d.Payload == other.Payload &&
		d.UpdateCounter == other.UpdateCounter &&
		d.Seed == other.Seed &&
		d.Key == other.Key
}

func generateKeyFromSeed(seed int64) string {
	return strconv.FormatInt(seed, 10)
}

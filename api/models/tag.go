package models

func MapToTags(tags map[string]string) *[]Tag {
	ret := make([]Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, Tag{
			Key:   key,
			Value: val,
		})
	}
	return &ret
}

func MergeTags(left, right *[]Tag) *[]Tag {
	if left == nil && right == nil {
		return nil
	}

	merged := &[]Tag{}
	if left == nil || len(*left) == 0 {
		*merged = *right

		return merged
	}

	if right == nil || len(*right) == 0 {
		*merged = *left

		return merged
	}

	m := map[string]string{}
	for _, tag := range *left {
		m[tag.Key] = tag.Value
	}

	for _, tag := range *right {
		m[tag.Key] = tag.Value
	}

	return MapToTags(m)
}

package tags

import (
	"api/src/domain/tag"
)

type TagHandler struct {
	tagRepo tag.TagRepository
}

func NewTagHandler(tagRepo tag.TagRepository) *TagHandler {
	return &TagHandler{tagRepo: tagRepo}
}

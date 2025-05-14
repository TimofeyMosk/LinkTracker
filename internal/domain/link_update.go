package domain

type LinkUpdate struct {
	Link        Link
	TgIDs       []int64
	Description string
}

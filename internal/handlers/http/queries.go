package http

type EmptyQuery struct{}

type IDQuery struct {
	ID string `path:"id" validate:"required"`
}

type SlugQuery struct {
	Slug string `path:"slug" validate:"required"`
}

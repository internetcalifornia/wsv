package internal

import "errors"

var (
	ErrLineNotFound       = errors.New("line does not exist")
	ErrFieldNotFound      = errors.New("field does not exist")
	ErrOmitHeaders        = errors.New("document configured to omit headers")
	ErrInvalidPaddingRune = errors.New("only whitespace characters can be used for padding")
	ErrStartedToWrite     = errors.New("document started to write, need to reset document to edit")
	ErrLineIsNotHeader    = errors.New("the line is not the first line in the document")
	ErrNotEnoughLines     = errors.New("document does not have more than 1 line")
	ErrLineNotEditable    = errors.New("cannot edit")
	ErrFieldCount         = errors.New("wrong number of fields")
	ErrBareQuote          = errors.New("bare \" in non-quoted-field")
	ErrQuote              = errors.New("extraneous or missing \" in quoted-field")
	ErrLineFeedTerm       = errors.New("line feed terminated before the line end end")
	ErrInvalidNull        = errors.New("null `-` specifier cannot be included without white space surrounding, unless it is the last value in the line. To record a literal `-` please wrap the value in double quotes")
	ErrCommentPlacement   = errors.New("comments should be the last elements in a row, if immediate preceding lines are null, they cannot be omitted and must be explicitly declared")
	ErrReaderEnded        = errors.New("reader ended, nothing left to read")
)

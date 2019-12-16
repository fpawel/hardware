package ktx500

import "github.com/ansel1/merry"

var (
	Err = merry.New("КТХ-500")
)

func wrapErr(err error) merry.Error {
	if merry.Is(err, Err) {
		return merry.Wrap(err)
	}
	return merry.WithCause(err, Err)
}

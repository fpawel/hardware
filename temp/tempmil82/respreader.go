package tempmil82

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/hardware/internal/pkg"
)

func NewResponseReader(p *comport.Port, c comm.Config) ResponseReader {
	return responseReader{
		p:   p,
		cfg: c,
	}
}

type responseReader struct {
	p   *comport.Port
	cfg comm.Config
}

func (x responseReader) GetResponse(log comm.Logger, ctx context.Context, request []byte) ([]byte, error) {
	log = pkg.LogPrependSuffixKeys(log, "comport", x.p.Config().Name)
	b, err := comm.NewResponseReader(ctx, x.p, x.cfg, nil).GetResponse(request, log)
	return b, merry.Appendf(err, "comport=%s", x.p.Config().Name)
}

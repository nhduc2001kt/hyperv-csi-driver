package hypervwinrmimpl

import (
	"context"
	_ "embed"
	"text/template"

	"github.com/nhduc2001kt/hyperv-csi-driver/pkg/hyperv"
)

var (
	//go:embed scripts/Get-VMByID.ps1
	getVMByIDFile string
)

var (
	getVmTemplate = template.Must(template.New("GetVMByID").Parse(getVMByIDFile))
)

type getVMByIDArgs struct {
	ID string
}

func (c *hypervClientImpl) GetVMByID(ctx context.Context, id string) (result hyperv.VM, err error) {
	err = c.winrmClient.RunScriptWithResult(ctx, getVmTemplate, getVMByIDArgs{
		ID: id,
	}, &result)

	return result, err
}

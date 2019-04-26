package cmd

import (
	"context"
	"fmt"
	"github.com/ribencong/go-youPipe/core"
	"github.com/ribencong/go-youPipe/pbs"
	"github.com/spf13/cobra"
)

var adminParam = struct {
	maxBootNode int
}{}

/************************************************************************
*							Client part
************************************************************************/
func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.Flags().IntVarP(&adminParam.maxBootNode, "maxBoot", "m", 8, "Max boot strap number to select")
}

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "This is for semi-blockChain administrator usage",
	Long:  `TODO::.`,
	Run:   admin,
	Args:  cobra.MinimumNArgs(1),
}

func admin(_ *cobra.Command, args []string) {
	switch args[0] {
	case "fb":
		findSomeBootStrapNode()
	}
}

func findSomeBootStrapNode() {

	logger.Debugf("try to find (%d) boot strap node......", adminParam.maxBootNode)

	msg := &pbs.BootNodeReq{
		MaxSize: int32(adminParam.maxBootNode),
	}

	client := DialToCmdService()

	res, err := client.FindBootNode(context.Background(), msg)

	fmt.Println(res, err)
}

/************************************************************************
*							Service part
************************************************************************/

func (s *cmdService) FindBootNode(ctx context.Context, request *pbs.BootNodeReq) (*pbs.BootNodeRes, error) {

	path, _ := core.GetNodeInst().AdminFindBootStrap(request.MaxSize)

	return &pbs.BootNodeRes{Nodes: path}, nil
}

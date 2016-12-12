package polly

import (
	_ "errors"
	_ "flag"
	_ "os"
	_ "fmt"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

var (
	CLSID_FilterGraph = ole.NewGUID("{E436EBB3-524F-11CE-9F53-0020AF0BA770}")

	IID_IGraphBuilder = ole.NewGUID("{56A868A9-0AD4-11CE-B03A-0020AF0BA770}")

	IID_IMediaControl = ole.NewGUID("{56A868B1-0AD4-11CE-B03A-0020AF0BA770}")

	IID_IMediaEvent   = ole.NewGUID("{56a868b6-0ad4-11ce-b03a-0020af0ba770}")

	INFINITE = int(-1)
)

func Play(filename string) (err error) {
	connection := &ole.Connection{nil}

	err = connection.Initialize()
	if err != nil {
		return
	}
	defer connection.Uninitialize()

	pGraphBuilder, err := ole.CreateInstance(CLSID_FilterGraph, IID_IGraphBuilder)
	if err != nil {
		return
	}

	pMediaControl,err := pGraphBuilder.QueryInterface(IID_IMediaControl)
	if err != nil {
		return
	}

	pMediaEvent,err := pGraphBuilder.QueryInterface(IID_IMediaEvent)
	if err != nil {
		return
	}

	_ = oleutil.MustCallMethod(pMediaControl, "RenderFile", filename)

	_ = oleutil.MustCallMethod(pMediaControl, "Run")

	for ev := 0; ev == 0; {
		_ = oleutil.MustCallMethod(pMediaEvent, "WaitForCompletion", INFINITE, &ev)
	}
	return
}

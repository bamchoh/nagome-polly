package player

import (
	"io/ioutil"
	"os"
	"io"
	"bufio"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/aws/aws-sdk-go/service/polly"
)

var (
	CLSID_FilterGraph = ole.NewGUID("{E436EBB3-524F-11CE-9F53-0020AF0BA770}")

	IID_IGraphBuilder = ole.NewGUID("{56A868A9-0AD4-11CE-B03A-0020AF0BA770}")

	IID_IMediaControl = ole.NewGUID("{56A868B1-0AD4-11CE-B03A-0020AF0BA770}")

	IID_IMediaEvent   = ole.NewGUID("{56a868b6-0ad4-11ce-b03a-0020af0ba770}")

	INFINITE = int(-1)
)

func save_file(resp *polly.SynthesizeSpeechOutput, f *os.File) {
	writer := bufio.NewWriter(f)
	io.Copy(writer, resp.AudioStream)
	writer.Flush()
}


func Play(resp *polly.SynthesizeSpeechOutput) error {
	tmpfile,err := ioutil.TempFile("","ngm-polly-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	save_file(resp, tmpfile)
	if err := tmpfile.Close(); err != nil {
		return err
	}

	return play(tmpfile.Name())
}


func play(filename string) (err error) {
	connection := &ole.Connection{nil}

	err = connection.Initialize()
	if err != nil {
		return
	}
	defer connection.Uninitialize()

	pGraphBuilder, err := ole.CreateInstance(CLSID_FilterGraph, IID_IGraphBuilder)
	if err != nil {
		if pGraphBuilder != nil {
			pGraphBuilder.Release()
		}
		ole.CoUninitialize()
		return
	}

	pMediaControl,err := pGraphBuilder.QueryInterface(IID_IMediaControl)
	if err != nil {
		if pMediaControl != nil {
			pMediaControl.Release()
		}

		if pGraphBuilder != nil {
			pGraphBuilder.Release()
		}
		ole.CoUninitialize()
		return
	}

	pMediaEvent,err := pGraphBuilder.QueryInterface(IID_IMediaEvent)
	if err != nil {
		if pMediaEvent != nil {
			pMediaEvent.Release()
		}

		if pMediaControl != nil {
			pMediaControl.Release()
		}

		if pGraphBuilder != nil {
			pGraphBuilder.Release()
		}
		ole.CoUninitialize()
		return
	}

	var res *ole.VARIANT
	res = oleutil.MustCallMethod(pMediaControl, "RenderFile", filename)
	if res == nil {
		pMediaEvent.Release()
		pMediaControl.Release()
		pGraphBuilder.Release()
		ole.CoUninitialize()
		return
	}

	res = oleutil.MustCallMethod(pMediaControl, "Run")
	if res == nil {
		pMediaEvent.Release()
		pMediaControl.Release()
		pGraphBuilder.Release()
		ole.CoUninitialize()
		return
	}


	for ev := 0; ev == 0; {
		res = oleutil.MustCallMethod(pMediaEvent, "WaitForCompletion", INFINITE, &ev)
		if res == nil {
			pMediaEvent.Release()
			pMediaControl.Release()
			pGraphBuilder.Release()
			ole.CoUninitialize()
			return
		}
	}

	pMediaEvent.Release()
	pMediaControl.Release()
	pGraphBuilder.Release()
	ole.CoUninitialize()

	return
}

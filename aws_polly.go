package main

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
)

type PollyConfig struct {
	Region   string
	Format   string
	Voice    string
	TextType string
	Text     string
	Polly    *polly.Polly
}

func init_polly(region string) (*polly.Polly,error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil,err
	}
	return polly.New(sess, aws.NewConfig().WithRegion(region)),nil

}

func synthesize_speech(pc PollyConfig) (*polly.SynthesizeSpeechOutput,error) {
	params := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(pc.Format),
		Text:         aws.String(pc.Text),
		TextType:     aws.String(pc.TextType),
		VoiceId:      aws.String(pc.Voice),
	}
	return pc.Polly.SynthesizeSpeech(params)
}


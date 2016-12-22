package main

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/polly"

	"gopkg.in/yaml.v2"
	"io/ioutil"
	"io"
	"errors"
)

type PollyConfig struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
	Format    string `yaml:"format"`
	Voice     string `yaml:"voice"`
	TextType  string `yaml:"text_type"`
	Polly     *polly.Polly
}

func load(r io.Reader) (*PollyConfig, error) {
	var data []byte
	var err error
	var pc PollyConfig

	data,err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &pc)
	if err != nil {
		return nil, err
	}

	err_txt := ""
	if pc.AccessKey == "" {
		err_txt += "value of access_key does not set in setting file. "
	}

	if pc.SecretKey == "" {
		err_txt += "value of secret_key does not set in setting file. "
	}

	if err_txt != "" {
		return nil, errors.New(err_txt)
	}

	if pc.Region == "" {
		pc.Region = "us-west-2"
	}

	if pc.Format == "" {
		pc.Format = "mp3"
	}

	if pc.Voice == "" {
		pc.Voice = "Mizuki"
	}

	if pc.TextType == "" {
		pc.TextType = "ssml"
	}

	pc.Polly, err = init_polly(&pc)
	if err != nil {
		return nil, err
	}

	return &pc, err
}

func init_polly(pc *PollyConfig) (*polly.Polly,error) {
	creds := credentials.NewStaticCredentials(pc.AccessKey, pc.SecretKey, "")
	sess := session.New(&aws.Config{Credentials: creds})
	return polly.New(sess, aws.NewConfig().WithRegion(pc.Region)),nil
}

func synthesize_speech(pc *PollyConfig, text string) (*polly.SynthesizeSpeechOutput,error) {
	params := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(pc.Format),
		Text:         aws.String(text),
		TextType:     aws.String(pc.TextType),
		VoiceId:      aws.String(pc.Voice),
	}
	return pc.Polly.SynthesizeSpeech(params)
}


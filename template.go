package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"go.uber.org/zap"
)

type TemplateData struct {
	Title   string
	Content string
	Foot    string
}

type FeishuRobot struct {
	log *zap.Logger

	url                string
	infoTmpl, warnTmpl *template.Template
}

func NewFeishuRobot(log *zap.Logger, u string) (*FeishuRobot, error) {
	var err error
	infoTmpl, err := template.ParseFiles("feishu_msg_info.tmpl")
	if err != nil {
		return nil, err
	}

	warnTmpl, err := template.ParseFiles("feishu_msg_warn.tmpl")
	if err != nil {
		return nil, err
	}
	return &FeishuRobot{
		log:      log,
		url:      u,
		infoTmpl: infoTmpl,
		warnTmpl: warnTmpl,
	}, nil
}

func (r *FeishuRobot) SendInfoMessage(data TemplateData) error {
	var doc bytes.Buffer
	err := r.infoTmpl.Execute(&doc, data)
	if err != nil {
		return err
	}
	return r.sendMessage(bytes.ReplaceAll(doc.Bytes(), []byte("\n"), []byte("")))
}

func (r *FeishuRobot) SendWarnMessage(data TemplateData) error {
	var doc bytes.Buffer
	err := r.warnTmpl.Execute(&doc, data)
	if err != nil {
		return err
	}
	return r.sendMessage(bytes.ReplaceAll(doc.Bytes(), []byte("\n"), []byte("")))
}

func (r *FeishuRobot) sendMessage(b []byte) error {
	client := &http.Client{
		Timeout: time.Second * 3,
	}

	r.log.Debug("sendMessage", zap.String("body", string(b)))
	body := bytes.NewBuffer(b)
	request, err := http.NewRequest("POST", r.url, body)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("%d: %s", response.StatusCode, string(responseBody))
	}

	return nil
}

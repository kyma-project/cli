// Copied from:
// https://github.com/gboddin/go-www-authenticate-parser/commit/ec0b649bb07701564d8bf002d0262a7bc4c13fe8

package cis

import (
	"bytes"
	"encoding/json"
)

type wwwAuthenticateSettings struct {
	state        func() error
	digestBuffer *bytes.Buffer
	buffer       string
	currentParam string
	quoteOpened  bool
	AuthType     string
	Params       map[string]string
}

func ParseAuthSettings(digestBuffer string) wwwAuthenticateSettings {
	digest := wwwAuthenticateSettings{
		digestBuffer: bytes.NewBufferString(digestBuffer),
		buffer:       "",
		AuthType:     "",
		Params:       make(map[string]string),
	}
	digest.state = digest.ParseType
	for {
		if err := digest.state(); err != nil {
			break
		}
	}
	return digest
}

func (d *wwwAuthenticateSettings) ParseType() error {
	currentByte, err := d.digestBuffer.ReadByte()
	if err != nil {
		return err
	}
	if currentByte != ' ' && currentByte != '\n' {
		d.buffer += string(currentByte)
		return nil
	}
	d.AuthType = d.buffer
	d.buffer = ""
	d.state = d.ParseParamKey
	return nil
}

func (d *wwwAuthenticateSettings) ParseParamKey() error {
	currentByte, err := d.digestBuffer.ReadByte()
	if err != nil {
		return err
	}
	switch currentByte {
	case '=':
		d.currentParam = d.buffer
		d.buffer = ""
		d.state = d.ParseParamValue
		return nil
	case ' ':
		if len(d.buffer) > 0 {
			d.Params[d.buffer] = "true"
			d.currentParam = ""
			d.buffer = ""
			d.state = d.ParseParamKey
		}
		return nil
	case ',':
		if len(d.buffer) > 0 {
			d.Params[d.buffer] = "true"
		}
		d.currentParam = ""
		d.buffer = ""
		d.state = d.ParseParamKey
		return nil
	}
	d.buffer += string(currentByte)
	return nil
}

func (d *wwwAuthenticateSettings) ParseParamValue() error {
	currentByte, err := d.digestBuffer.ReadByte()
	if err != nil {
		return err
	}
	switch currentByte {
	case '\\':
		nextByte, err := d.digestBuffer.ReadByte()
		if err != nil {
			return err
		}
		var unquoted string
		err = json.Unmarshal([]byte("\""+string(currentByte)+string(nextByte)+"\""), &unquoted)
		if err != nil {
			return err
		}
		d.buffer += unquoted
		return nil
	case '"':
		if d.quoteOpened {
			d.quoteOpened = false
			d.Params[d.currentParam] = d.buffer
			d.currentParam = ""
			d.buffer = ""
			// Read until the next ','
			_, err = d.digestBuffer.ReadString(',')
			if err != nil {
				return err
			}
			d.state = d.ParseParamKey
			return nil
		}
		if !d.quoteOpened {
			d.quoteOpened = true
			return nil
		}
	case ',':
		if !d.quoteOpened {
			d.quoteOpened = false
			d.Params[d.currentParam] = d.buffer
			d.currentParam = ""
			d.buffer = ""
			d.state = d.ParseParamKey
			return nil
		}
	}

	d.buffer += string(currentByte)
	return nil
}

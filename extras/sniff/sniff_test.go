package sniff

import (
	"encoding/base64"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnifferTCP(t *testing.T) {
	sniffer := &Sniffer{
		Timeout:       1 * time.Second,
		RewriteDomain: false,
	}

	buf := &[]byte{}

	// Test HTTP
	*buf = []byte("POST /hello HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"User-Agent: mamamiya\r\n" +
		"Content-Length: 27\r\n" +
		"Connection: keep-alive\r\n\r\n" +
		"param1=value1&param2=value2")
	index := 0
	stream := &mockStream{}
	stream.EXPECT().SetReadDeadline(mock.Anything).Return(nil)
	stream.EXPECT().Read(mock.Anything).RunAndReturn(func(bs []byte) (int, error) {
		if index < len(*buf) {
			n := copy(bs, (*buf)[index:])
			index += n
			return n, nil
		} else {
			return 0, io.EOF
		}
	})

	// Rewrite IP to domain
	reqAddr := "111.111.111.111:80"
	assert.True(t, sniffer.Check(false, reqAddr))
	putback, err := sniffer.TCP(stream, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, *buf, putback)
	assert.Equal(t, "example.com:80", reqAddr)

	// Do not rewrite if it's already a domain
	index = 0
	reqAddr = "gulag.cc:443"
	assert.False(t, sniffer.Check(false, reqAddr))

	// Turn on rewrite and now it should rewrite
	sniffer.RewriteDomain = true
	assert.True(t, sniffer.Check(false, reqAddr))

	// Test TLS
	*buf, err = base64.StdEncoding.DecodeString("FgMBARcBAAETAwPJL2jlt1OAo+Rslkjv/aqKiTthKMaCKg2Gvd+uALDbDCDdY+UIk8ouadEB9fC3j52Y1i7SJZqGIgBRIS6kKieYrAAoEwITAcAswCvAMMAvwCTAI8AowCfACsAJwBTAEwCdAJwAPQA8ADUALwEAAKIAAAAOAAwAAAlpcGluZm8uaW8ABQAFAQAAAAAAKwAJCAMEAwMDAgMBAA0AGgAYCAQIBQgGBAEFAQIBBAMFAwIDAgIGAQYDACMAAAAKAAgABgAdABcAGAAQAAsACQhodHRwLzEuMQAzACYAJAAdACBguQbqNJNyamYxYcrBFpBP7pWv5TgZsP9gwGtMYNKVBQAxAAAAFwAA/wEAAQAALQACAQE=")
	assert.NoError(t, err)
	index = 0
	reqAddr = "222.222.222.222:443"
	assert.True(t, sniffer.Check(false, reqAddr))
	putback, err = sniffer.TCP(stream, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, *buf, putback)
	assert.Equal(t, "ipinfo.io:443", reqAddr)

	// Test unrecognized 1
	*buf = []byte("Wait It's All Ohio? Always Has Been.")
	index = 0
	reqAddr = "123.123.123.123:123"
	assert.True(t, sniffer.Check(false, reqAddr))
	putback, err = sniffer.TCP(stream, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, *buf, putback)
	assert.Equal(t, "123.123.123.123:123", reqAddr)

	// Test unrecognized 2
	*buf = []byte("\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a")
	index = 0
	reqAddr = "45.45.45.45:45"
	assert.True(t, sniffer.Check(false, reqAddr))
	putback, err = sniffer.TCP(stream, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, []byte("\x01\x02\x03"), putback)
	assert.Equal(t, "45.45.45.45:45", reqAddr)

	// Test timeout
	blockStream := &mockStream{}
	blockStream.EXPECT().SetReadDeadline(mock.Anything).Return(nil)
	blockStream.EXPECT().Read(mock.Anything).RunAndReturn(func(bs []byte) (int, error) {
		time.Sleep(2 * time.Second)
		return 0, io.EOF
	})
	reqAddr = "66.66.66.66:66"
	assert.True(t, sniffer.Check(false, reqAddr))
	putback, err = sniffer.TCP(blockStream, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, putback)
	assert.Equal(t, "66.66.66.66:66", reqAddr)
}

func TestSnifferUDP(t *testing.T) {
	sniffer := &Sniffer{
		Timeout:       1 * time.Second,
		RewriteDomain: false,
	}

	// Test QUIC
	reqAddr := "2.3.4.5:443"
	assert.True(t, sniffer.Check(true, reqAddr))
	pkt, err := base64.StdEncoding.DecodeString("ygAAAAEIwugWgPS7ulYAAES8hY891uwgGE9GG4CPOLd+nsDe28raso24lCSFmlFwYQG1uF39ikbL13/R9ZTghYmTl+jEbr6F9TxxRiOgpTmKRmh6aKZiIiVfy5pVRckovaI8lq0WRoW9xoFNTyYtQP8TVJ3bLCK+zUqpquEQSyWf7CE43ywayyMpE9UlIoPXFWCoopXLM1SvzdQ+17P51N9KR7m4emti4DWWTBLMQOvrwd2HEEkbiZdRO1wf6ZXJlIat5dN0R/6uod60OFPO+u+awvq67MoMReC7+5I/xWI+xx6o4JpnZNn6YPG8Gqi8hS6doNcAAdtD8h5eMLuHCCgkpX3QVjjfWtcOhtw9xKjU43HhUPwzUTv+JDLgwuTQCTmlfYlb3B+pk4b2I9si0tJ0SBuYaZ2VQPtZbj2hpGXw3gn11pbN8xsbKkQL50+Scd4dGJxWQlGaJHeaU5WOCkxLXc635z8m5XO/CBHVYPGp4pfwfwNUgbe5WF+3MaUIlDB8dMfsnrO0BmZPo379jVx0SFLTAiS8wAdHib1WNEY8qKYnTWuiyxYg1GZEhJt0nXmI+8f0eJq42DgHBWC+Rf5rRBr/Sf25o3mFAmTUaul0Woo9/CIrpT73B63N91xd9A77i4ru995YG8l9Hen+eLtpDU9Q9376nwMDYBzeYG9U/Rn0Urbm6q4hmAgV/xlNJ2rAyDS+yLnwqD6I0PRy8bZJEttcidb/SkOyrpgMiAzWeT+SO+c/k+Y8H0UTRa05faZUrhuUaym9wAcaIVRA6nFI+fejfjVp+7afFv+kWn3vCqQEij+CRHuxkltrixZMD2rfYj6NUW7TTYBtPRtuV/V0ZIDjRR26vr4K+0D84+l3c0mA/l6nmpP5kkco3nmpdjtQN6sGXL7+5o0nnsftX5d6/n5mLyEpP+AEDl1zk3iqkS62RsITwql6DMMoGbSDdUpMclCIeM0vlo3CkxGMO7QA9ruVeNddkL3EWMivl+uxO43sXEEqYQHVl4N75y63t05GOf7/gm9Kb/BJ8MpG9ViEkVYaskQCzi3D8bVpzo8FfTj8te8B6c3ikc/cm7r8k0ZcZpr+YiLGDYq+0ilHxpqJfmq8dPkSvxdzLcUSvy7+LMQ/TTobRSF7L4JhtDKck0+00vl9H35Tkh9N+MsVtpKdWyoqZ4XaK2Nx1M6AieczXpdFc0y7lYPoUfF4IeW8WzeVUclol5ElYjkyFz/lDOGAe1bF2g5AYaGWCPiGleVZknNdD5ihB8W8Mfkt1pEwq2S97AHrppqkf/VoIfZzeqH8wUFw8fDDrZIpnoa0rW7HfwIQaqJhPCyB9Z6TVbV4x9UWmaHfVAcinCK/7o10dtaj3rvEqcUC/iPceGq3Tqv/p9GGNJ+Ci2JBjXqNxYr893Llk75VdPD9pM6y1SM0P80oXNy32VMtafkFFST8GpvvqWcxUJ93kzaY8RmU1g3XFOImSU2utU6+FUQ2Pn5uLwcfT2cTYfTpPGh+WXjSbZ6trqdEMEsLHybuPo2UN4WpVLXVQma3kSaHQggcLlEip8GhEUAy/xCb2eKqhI4HkDpDjwDnDVKufWlnRaOHf58cc8Woi+WT8JTOkHC+nBEG6fKRPHDG08U5yayIQIjI")
	assert.NoError(t, err)
	err = sniffer.UDP(pkt, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, "www.notion.so:443", reqAddr)

	// Test unrecognized
	pkt = []byte("oh my sweet summer child")
	reqAddr = "90.90.90.90:90"
	assert.True(t, sniffer.Check(true, reqAddr))
	err = sniffer.UDP(pkt, &reqAddr)
	assert.NoError(t, err)
	assert.Equal(t, "90.90.90.90:90", reqAddr)
}

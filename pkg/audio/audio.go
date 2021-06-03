package audio

import (
	"github.com/gen2brain/malgo"
)

type DeviceType uint32

const (
	Playback DeviceType = iota + 1
	Capture
	Duplex
	Loopback
)

type AudioCallback func(out, in []byte, samples int, samplesize int)

type AudioStream struct {
	ctx        *malgo.AllocatedContext
	cfg        malgo.DeviceConfig
	device     *malgo.Device
	callback   AudioCallback
	channels   int
	samplesize int
}

func (as *AudioStream) audioCallback(out, in []byte, framecount uint32) {
	var samples = int(framecount) * as.channels * as.samplesize
	if as.callback != nil {
		as.callback(out, in, samples, as.samplesize)
	}
}

func (as *AudioStream) SetCallback(callback AudioCallback) {
	as.callback = callback
}

func (as *AudioStream) Start() {
	as.device.Start()
}

func (as *AudioStream) Stop() {
	as.device.Stop()
}

func (as *AudioStream) Close() {
	as.Stop()
	as.device.Uninit()
	_ = as.ctx.Uninit()
	as.ctx.Free()
}

func NewAudioStream(devicetype DeviceType, channels int, sampleRate int) (as *AudioStream, err error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		return
	}
	as = &AudioStream{
		ctx:      ctx,
		cfg:      malgo.DefaultDeviceConfig(malgo.DeviceType(devicetype)),
		callback: nil,
		channels: channels,
	}
	as.cfg.Capture.Format = malgo.FormatS16
	as.cfg.Capture.Channels = uint32(as.channels)
	as.cfg.Playback.Format = malgo.FormatS16
	as.cfg.Playback.Channels = uint32(as.channels)
	as.cfg.SampleRate = uint32(sampleRate)
	as.cfg.Alsa.NoMMap = 1
	as.samplesize = malgo.SampleSizeInBytes(malgo.FormatS16)
	as.device, err = malgo.InitDevice(ctx.Context, as.cfg, malgo.DeviceCallbacks{
		Data: as.audioCallback,
	})
	if err != nil {
		return
	}
	return
}

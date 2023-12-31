package audio

import (
	"runtime"

	"github.com/Kor-SVS/cocoa/src/audio/dsp"
	"github.com/Kor-SVS/cocoa/src/util"
	"github.com/gen2brain/malgo"
	"github.com/gopxl/beep"
)

type AudioStream struct {
	beep.StreamSeekCloser
	Format *beep.Format
}

var (
	audioStreamBroker *util.Broker[EnumAudioStreamState]
)

var (
	audioStream             *AudioStream                // 오디오 데이터 스트림
	bufferCallbackFuncArray []func(buffer [][2]float64) // 오디오 버퍼 접근 함수 콜백
	audioBuffer             [][2]float64                // 오디오 버퍼 (재생)

	isValidAudioReadBuffer bool      // 오디오 버퍼 유효 여부
	audioReadBuffer        []float64 // 오디오 버퍼 (버퍼 읽기)
)

func init() {
	audioStreamBroker = util.NewBroker[EnumAudioStreamState]()
	audioStreamBroker.Start()
}

func Open(fpath string) {
	audioMutex.Lock()
	defer audioMutex.Unlock()

	logger.Trace("오디오 파일을 여는 중...")

	decoder, format, err := GetDecoder(fpath)
	if err != nil {
		logger.Errorf("오디오 파일을 열지 못했습니다. (err=%v)", err)
		return
	}

	if decodeErr := decoder.Err(); decodeErr != nil {
		logger.Errorf("디코드 오류 발생 (decodeErr=%v)", decodeErr)
		return
	}

	disposeStream()

	audioStream = &AudioStream{}
	audioStream.StreamSeekCloser = decoder
	audioStream.Format = format
	audioBuffer = nil

	deviceConfig := newDeviceConfig()

	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = uint32(audioStream.Format.SampleRate)

	initDevice(deviceConfig)

	logger.Infof("오디오 로드 완료 (fpath=%v)", fpath)
	logger.Infof("오디오 정보 (format={sr=%v, n_channels=%v, precision=%v}, sampleCount=%v)",
		format.SampleRate,
		format.NumChannels,
		format.Precision,
		audioStream.Len(),
	)

	audioStreamBroker.Publish(EnumAudioStreamOpen)
}

func GetMonoAllSampleData() []float64 {
	audioMutex.Lock()
	defer audioMutex.Unlock()

	if isValidAudioReadBuffer {
		return audioReadBuffer
	} else {
		buf := readAllSampleData()
		audioReadBuffer = dsp.StereoToMono(buf)
		isValidAudioReadBuffer = true
	}

	return audioReadBuffer
}

func GetAllSampleData() [][2]float64 {
	audioMutex.Lock()
	defer audioMutex.Unlock()

	return readAllSampleData()
}

func readAllSampleData() [][2]float64 {
	buf := make([][2]float64, audioStream.Len())

	audioStreamReadMutex.Lock()
	pos := audioStream.Position()
	audioStream.Seek(0)

	audioStream.Stream(buf)

	audioStream.Seek(pos)
	audioStreamReadMutex.Unlock()

	return buf
}

func readAudioStream(outBuffer []byte, frameCount int) int {
	audioStreamReadMutex.Lock()
	defer audioStreamReadMutex.Unlock()

	if audioStream == nil {
		return 0
	}

	if audioBuffer == nil || len(audioBuffer) != frameCount {
		audioBuffer = make([][2]float64, frameCount)
	}

	sampleLen := audioStream.Len()
	readN, ok := audioStream.Stream(audioBuffer)

	for _, callback := range bufferCallbackFuncArray {
		callback(audioBuffer)
	}

	dsp.FloatSampleToByteArray(audioBuffer, outBuffer)

	if readN == sampleLen && ok { // 스트림이 끝났을 경우
		return readN
	}
	return readN
}

func isAudioLoaded() bool {
	return audioStream != nil && audioDevice != nil
}

func AudioStreamBroker() *util.Broker[EnumAudioStreamState] {
	return audioStreamBroker
}

func Close() {
	audioMutex.Lock()
	defer audioMutex.Unlock()

	disposeStream()
}

func disposeStream() {
	disposeDevice()

	if audioStream != nil {
		audioStream.Close()
		audioStream = nil
		audioBuffer = nil
		isValidAudioReadBuffer = false
		audioReadBuffer = nil
		audioStreamBroker.Publish(EnumAudioStreamClosed)
		runtime.GC()
	}
}

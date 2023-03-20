'use strict'
import {onMount} from "solid-js";

/********************************************************************
 created:    2023-03-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export default function () {
    // todo UI控制相关的，应该跟model层拆到两个地方
    onMount(() => {
        const testKey = 'Control'
        document.addEventListener('keydown', evt => {
            if (evt.key === testKey) {
                startRecording()
            }
        })

        document.addEventListener('keyup', evt => {
            if (evt.key === testKey) {
                stopRecording()
            }
        })
    })

    const kMaxAudioSeconds = 120
    const kSampleRate = 16000

    const audioContext = new AudioContext({
        sampleRate: kSampleRate,
        channelCount: 1,
        echoCancellation: false,
        autoGainControl: true,
        noiseSuppression: true,
    })

    let isRecording = false
    let mediaRecorder = undefined

    // record up to kMaxAudio_s seconds of audio from the microphone
    // check if doRecording is false every 1000 ms and stop recording if so
    // update progress information
    function startRecording() {
        if (isRecording) {
            return
        }

        isRecording = true
        console.log('start recording~')
        navigator.mediaDevices.getUserMedia({audio: true, video: false})
            .then(function (stream) {
                mediaRecorder = new MediaRecorder(stream)

                let chunks = []
                mediaRecorder.ondataavailable = function (e) {
                    chunks.push(e.data)
                }

                function resetData() {
                    chunks = []
                    stream.getTracks().forEach(function (track) {
                        track.stop()
                    })
                }

                // 录音结束的时候，会调用onstop，然后把chunks中的内容写到blob中，而后是使用reader读取blob中的内容，读成功后走到reader.onload()
                mediaRecorder.onstop = function (e) {
                    console.log('stop recording~')
                    const blob = new Blob(chunks, {'type': 'audio/ogg; codecs=opus'})
                    resetData()

                    const reader = new FileReader()
                    reader.onload = function (event) {
                        const data = reader.result
                        if (data.byteLength === 0) {
                            console.log('no audio data recorded')
                            return
                        }

                        const buf = new Uint8Array(data)
                        audioContext.decodeAudioData(buf.buffer, function (audioBuffer) {
                            const offlineContext = new OfflineAudioContext(audioBuffer.numberOfChannels, audioBuffer.length, audioBuffer.sampleRate)
                            const source = offlineContext.createBufferSource()
                            source.buffer = audioBuffer
                            source.connect(offlineContext.destination)
                            source.start(0)

                            offlineContext.startRendering().then(function (renderedBuffer) {
                                let audio = renderedBuffer.getChannelData(0)
                                console.log('audio recorded, size: ' + audio.length)

                                // truncate to first 30 seconds
                                if (audio.length > kMaxAudioSeconds * kSampleRate) {
                                    audio = audio.slice(0, kMaxAudioSeconds * kSampleRate);
                                    console.log('truncated audio to first ' + kMaxAudioSeconds + ' seconds')
                                }
                                onSetAudio(audio)
                            })
                        }, function (e) {
                            console.error('error decoding audio: ' + e)
                            onSetAudio(undefined)
                        })
                    }

                    reader.readAsArrayBuffer(blob)
                }
                mediaRecorder.start()
            })
            .catch(function (err) {
                console.error('error getting audio stream: ' + err)
            })

        setTimeout(function () {
            if (isRecording) {
                console.log(`recording stopped after ${kMaxAudioSeconds} seconds`)
                stopRecording()
            }
        }, kMaxAudioSeconds * 1000)
    }

    function stopRecording() {
        if (isRecording) {
            isRecording = false
            mediaRecorder.stop()
        }
    }

    // 录音结束了，设置到这里
    function onSetAudio(audio) {
        console.log('audio set')
    }

    return <>
    </>
}
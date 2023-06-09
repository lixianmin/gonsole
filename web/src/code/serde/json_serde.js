/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function newJsonSerde() {
    const encoder = new TextEncoder()
    const decoder = new TextDecoder()

    function serialize(v) {
        const text = JSON.stringify(v)
        return encoder.encode(text)
    }

    function deserialize(bytes) {
        const text = decoder.decode(bytes)
        return JSON.parse(text)
    }

    return {
        serialize: serialize,
        deserialize: deserialize,
    }
}
/********************************************************************
 created:    2022-01-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {Buffers} from "./core/buffers";

/**
 * pomele client encode
 * id message id;
 * route message route
 * msg message body
 * socketio current support string
 */
export function strencode(str: string): Uint8Array {
    const buf = new ArrayBuffer(str.length * 3);
    const byteArray = new Uint8Array(buf);

    let offset = 0;
    for (let i = 0; i < str.length; i++) {
        const charCode = str.charCodeAt(i);

        let codes: number[];
        if (charCode <= 0x7f) {
            codes = [charCode];
        } else if (charCode <= 0x7ff) {
            codes = [0xc0 | (charCode >> 6), 0x80 | (charCode & 0x3f)];
        } else {
            codes = [0xe0 | (charCode >> 12), 0x80 | ((charCode & 0xfc0) >> 6), 0x80 | (charCode & 0x3f)];
        }

        for (let j = 0; j < codes.length; j++) {
            byteArray[offset] = codes[j];
            ++offset;
        }
    }

    const result = new Uint8Array(offset)
    Buffers.blockCopy(byteArray, 0, result, 0, offset)
    return result;
}

/**
 * client decode
 * msg String data
 * return Message Object
 */
export function strdecode(buffer): string {
    const bytes = new Uint8Array(buffer);
    const array: Array<number> = [];
    let offset = 0;
    let charCode = 0;
    const end = bytes.length;

    while (offset < end) {
        if (bytes[offset] < 128) {
            charCode = bytes[offset];
            offset += 1;
        } else if (bytes[offset] < 224) {
            charCode = ((bytes[offset] & 0x3f) << 6) + (bytes[offset + 1] & 0x3f);
            offset += 2;
        } else {
            charCode = ((bytes[offset] & 0x0f) << 12) + ((bytes[offset + 1] & 0x3f) << 6) + (bytes[offset + 2] & 0x3f);
            offset += 3;
        }
        array.push(charCode)
    }

    let text = arrayToString(array)
    // console.log("buffer=", buffer, "array=", array, ", text=", text)
    return text
}

// 解决 String.fromCharCode.apply(null, ascii)) 报 Uncaught RangeError: Maximum call stack size exceeded 的问题
// https://stackoverflow.com/questions/12710001/how-to-convert-uint8-array-to-base64-encoded-string/12713326#12713326
function arrayToString(u8a: Array<number>): string {
    const CHUNK_SZ = 0x8000;
    const c: Array<string> = [];
    for (let i = 0; i < u8a.length; i += CHUNK_SZ) {
        c.push(String.fromCharCode.apply(null, u8a.slice(i, i + CHUNK_SZ)));
    }

    return c.join("");
}


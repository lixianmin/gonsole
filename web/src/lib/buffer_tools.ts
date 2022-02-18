/********************************************************************
 created:    2022-01-17
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class BufferTools {
    public static blockCopy(src: Uint8Array, srcOffset: number, dst: Uint8Array, dstOffset: number, count: number) {
        for (let index = 0; index < count; index++) {
            dst[dstOffset++] = src[srcOffset++]
        }
    }
}
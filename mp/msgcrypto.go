package mp

import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/binary"
    "fmt"
)

const BLOCK_SIZE = 32

func encryptMsg(random, msg, appId, aesKey []byte) []byte {
    msgLen := len(msg)
    textLen := 20+msgLen+len(appId)
    padNum := BLOCK_SIZE - (textLen % BLOCK_SIZE)
    if padNum == 0 {
        padNum = BLOCK_SIZE
    }
    textLen += padNum

    text := make([]byte, textLen)

    copy(text[:16], random)
    binary.BigEndian.PutUint32(text[16:20], uint32(msgLen))
    copy(text[20:], msg)
    copy(text[20+msgLen:], appId)

    pad := byte(padNum)
    for i := textLen-padNum; i < textLen; i++ {
        text[i] = pad
    }

    block, err := aes.NewCipher(aesKey)
    if err != nil {
        panic(err)
    }
    mode := cipher.NewCBCEncrypter(block, aesKey[:16])
    mode.CryptBlocks(text, text)
    return text
}

func decryptMsg(ciphertext, aesKey []byte) (random, msg, appId []byte, err error) {
    if len(ciphertext) < BLOCK_SIZE {
        err = fmt.Errorf("ciphertext length is too short: %d", len(ciphertext))
        return
    }
    if len(ciphertext) % BLOCK_SIZE != 0 {
        err = fmt.Errorf("ciphertext length is invalid: %d", len(ciphertext))
        return
    }

    text := make([]byte, len(ciphertext))

    block, err := aes.NewCipher(aesKey)
    if err != nil {
        panic(err)
    }
    mode := cipher.NewCBCDecrypter(block, aesKey[:16])
    mode.CryptBlocks(text, ciphertext)

    padNum := int(text[len(text)-1])
    if padNum < 1 || padNum > BLOCK_SIZE {
        err = fmt.Errorf("incorrect pad bytes num: %d", padNum)
        return
    }

    text = text[:len(text)-padNum]

    if len(text) < 20 {
        err = fmt.Errorf("")
    }
    random = text[:16]
    msgLen := int(binary.BigEndian.Uint32(text[16:20]))

    if len(text) <= 20+msgLen {
        err = fmt.Errorf("incorrect msg length: %d", msgLen)
        return
    }
    msg = text[20:20+msgLen]
    appId = text[20+msgLen:]
    return
}

package main

import (
    "crypto/sha256"
    "database/sql"
    "encoding/binary"
    "encoding/hex"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/chacha20"
    "math/big"
    //"time"
)

const verbose bool = true;
var num3072_prime = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 3072), big.NewInt(1103717))

func hashToStr(bytes [32]byte) (string) {
    for i, j := 0, 31; i < j; i, j = i+1, j-1 {
        bytes[i], bytes[j] = bytes[j], bytes[i]
    }
    return fmt.Sprintf("%x", bytes)
}

func main() {
    db, err := sql.Open("sqlite3", "file:/home/honeybadger/.bitcoin/signet/utxo.sqlite")
    if err != nil { panic(err) }
    defer db.Close()

    // TODO: read metadata
    //fmt.Printf("UTXO Snapshot at block %s, contains %d coins\n",
    //           hashToStr(blockHash), numUTXOs)

    rows, err := db.Query("SELECT * FROM utxos")
    if err != nil { panic(err) }
    defer rows.Close()

    //t := time.Now()

    num3072 := big.NewInt(1)

    for rows.Next() {
        var txid_hex string
        var vout uint64
        var value uint64
        var coinbase uint64
        var height uint64
        var scriptpubkey_hex string

        err = rows.Scan(&txid_hex, &vout, &value, &coinbase, &height, &scriptpubkey_hex)
        if err != nil { panic(err) }

        txid, err := hex.DecodeString(txid_hex)
        if err != nil { panic(err) }
        for i, j := 0, 31; i < j; i, j = i+1, j-1 {
            txid[i], txid[j] = txid[j], txid[i]
        }
        scriptpubkey, err := hex.DecodeString(scriptpubkey_hex)
        if err != nil { panic(err) }

        if verbose {
            fmt.Printf("\ttxid = %x\n", txid)
            fmt.Printf("\tvout = %d\n", vout)
            fmt.Printf("\tvalue = %d sats\n", value)
            fmt.Printf("\tcoinbase = %d\n", coinbase)
            fmt.Printf("\theight = %d\n", height)
            fmt.Printf("\tscriptPubKey = %x\n", scriptpubkey)
            fmt.Printf("\n")
        }

        //////////////////////////////////////
        // serialize to be ready for muhash
        //////////////////////////////////////
        muser := make([]byte, 0, 128)
        muser = append(muser, txid...)

        tmp4 := make([]byte, 4)
        tmp8 := make([]byte, 8)
        binary.LittleEndian.PutUint32(tmp4, uint32(vout))
        muser = append(muser, tmp4...)

        binary.LittleEndian.PutUint32(tmp4, 2*uint32(height) + uint32(coinbase))
        muser = append(muser, tmp4...)

        binary.LittleEndian.PutUint64(tmp8, value)
        muser = append(muser, tmp8...)

        // TODO: also handle larger pubkeyscript-sizes (compact size...)
        if len(scriptpubkey) > 250 {
            panic("TODO: implement compact size serialization, len of scriptPubKey is too long...")
        }
        muser = append(muser, byte(len(scriptpubkey)))
        muser = append(muser, scriptpubkey...)

        fmt.Printf("FIRST UTXO for muhash, serialized: %x\n", muser)

        ///////////////// next step, hash of that thing /////////////////
        hashed := sha256.Sum256(muser)
        fmt.Printf("SHA256 of the serialized UTXO: %x\n", hashed)

        //// chacha20 ////
        nonce := [12]byte{0,0,0,0,0,0,0,0,0,0,0,0}
        cc20, err := chacha20.NewUnauthenticatedCipher(hashed[:], nonce[:])
        if err != nil { panic(err) }
        var num3072_raw [384]byte
        cc20.XORKeyStream(num3072_raw[:], num3072_raw[:])
        fmt.Printf("Chacha20 of SHA256 of the serialized UTXO: %x\n", num3072_raw)

        for i, j := 0, 383; i < j; i, j = i+1, j-1 {
            num3072_raw[i], num3072_raw[j] = num3072_raw[j], num3072_raw[i]
        }

        num3072_insert := new(big.Int).SetBytes(num3072_raw[:])
        num3072_insert.Mod(num3072_insert, num3072_prime)

        num3072.Mul(num3072, num3072_insert)
        num3072.Mod(num3072, num3072_prime)

        var xxx [384]byte
        num3072.FillBytes(xxx[:])
        for _, b := range xxx {
            fmt.Printf("%02x", b)
        }
        for i, j := 0, 383; i < j; i, j = i+1, j-1 {
            xxx[i], xxx[j] = xxx[j], xxx[i]
        }
        fmt.Printf("\n")
        hashed = sha256.Sum256(xxx[:])
        fmt.Printf("Final SHA256 of the Num3072 (MuHash): %s\n", hashToStr(hashed))
    }
}
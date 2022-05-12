# UTXO dump tools
This repository contains various converting and processing tools for dumps of
the Bitcoin UTXO set, as generated by Bitcoin Core via the `dumptxoutset` RPC.
It was created in the course of reviewing and testing pull request #24952
(https://github.com/bitcoin/bitcoin/pull/24952), which adds a new option to dump
the UTXO set in SQLite database format.

## Overview of tools

### `utxo_to_sqlite`

`utxo_to_sqlite` is a simple tool for converting a compact-serialized UTXO set
to a SQLite database. A table `utxos` is created with the following schema
(matching PR #24952):
```
CREATE TABLE utxos(txid TEXT, vout INT, value INT, coinbase INT, height INT, scriptpubkey TEXT)
```

Run via:
```
$ cd utxo_to_sqlite
$ go run utxo_to_sqlite ~/.bitcoin/utxos.dat ./utxos.sqlite
```

Note that the first run likely takes longer, as golang has to fetch and build
the SQLite library (https://github.com/mattn/go-sqlite3) first.

### `calc_utxo_hash`
`calc_utxo_hash` calculates the UTXO set hash of a dump in SQLite format. Right
now only the _MuHash_ type is supported, which can be determined on the Bitcoin
Core node via the `gettxoutsetinfo` RPC (pass `"muhash"` as first parameter).

Run via:
```
$ cd calc_utxo_hash
$ go run calc_utxo_hash ./utxos.sqlite
```

## TODO
- support also calculating the compact-serialized hash

## Acknowledgment
This work has been supported by a Brink grant (https://brink.dev/).

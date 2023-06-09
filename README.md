# Experiment: New Format for Wikidata Dumps?

This is an experiment for a simpler, smaller and faster (to decompress)
data format for [Wikidata dumps](https://www.wikidata.org/wiki/Wikidata:Database_download).

| Format      |     Size¹ | Tool   | Decompression time² |
| ----------- | --------: | ------ | ------------------: |
| `.qs.zst`   |  26.6 GiB | zstd   |           6 minutes |
| `.qs.bz2`   |  26.3 GiB | lbzip2 |           7 minutes |
| `.qs.br`    |  25.9 GiB | brotli |          11 minutes |
| `.qs.gz`    |  38.9 GiB | gzip   |          23 minutes |
| `.json.zst` |  72.3 GiB | zstd   |          19 minutes |
| `.json.br`  |  67.8 GiB | brotli |          25 minutes |
| `.json.bz2` |  75.9 GiB | lbzip2 |          59 minutes |
| `.json.gz`  | 115.2 GiB | gzip   |          89 minutes |
| `.qs.bz2`   |  26.3 GiB | pbzip2 |          94 minutes |
| `.json.bz2` |  75.9 GiB | pbzip2 |         926 minutes |

The proposed new format,
[QuickStatements](https://www.wikidata.org/wiki/Help:QuickStatements)
with [Zstandard](https://en.wikipedia.org/wiki/Zstd) compression,
takes about a third of the current best file size. On a typical modern
cloud server, decompression is about 10 times faster compared to lbzip2
on the current JSON dumps.


## Motivation

As of May 2023, the most compact format for [Wikidata
dumps](https://dumps.wikimedia.org/wikidatawiki/entities/20230424/) is
JSON with bzip2 compression.  However, the current JSON syntax is very
verbose, which makes it slow to process. Another issue is bzip2: since
its invention 27 years ago, newer algorithms have been designed for
fast decompression on today’s machines.

As a frequent user of Wikidata dumps, I got annoyed by the high cost of
processing the current format, and I wondered how much could be gained
from a better format. Specifically, a new format should be significantly
smaller in size; much faster to decompress; and easy to understand.

Wikidata editors frequently use the [QuickStatements
tool](https://www.wikidata.org/wiki/Help:QuickStatements) for bulk
maintenance. The tool accepts statements in a text syntax that is easy
to understand and quite compact. I wondered if Wikidata dumps could be
encoded as QuickStatements, and compressed with a modern algorithm
such as [Zstandard](https://en.wikipedia.org/wiki/Zstd).


## Extensions to QuickStatements syntax

Note that the current QuickStatements syntax cannot express all of
Wikidata; the major missing piece is ranking. For this experiment, I
encocded preferred and deprecated rank with ↑ and ↓ arrows, as in
`Q12|P9|↑"foo"`. All other missing parts are minor and rare, such as
coordinates on Venus and Mars; for this experiment, I pretended these
were on Earth. To fully encode all of Wikidata as QuickStatements,
suitable syntax would need to be defined and properly documented.
Obviously, it would then also make sense to support this new syntax
in the live QuickStatments tool.

Currently, QuickStatements does not seem to define an escape mechanism
for quote characters. In my experiment, I used an Unicode escape sequence
when a quoted string contained a quote, as in `"foo \u0022 bar"`.

A nice property of the current JSON format is that each item is encoded
on a separate line. It might be nice to preserve this property. This would
need (small, backwards-compatible) extensions to the QuickStatement syntax:
(a) allow multiple labels, aliases
and sitelinks, as in `Q2|Len|"Earth"|Aen|"Planet Earth"|Lfr|"Terre"`;
(b) allow multiple claims (not just multiple qualifiers) on the same
line, perhaps with a `!P` construct similar to the existing `!S`.
This would also make the format slightly more compact.

## Other issues with Wikidata dumps

In a new version of Wikidata dumps, I think it would be good to
address some other things.

1. Wikidata dumps should be atomic snapshots, taken at a defined point
in time. Currently, each item is getting dumped at a slightly different
time. This fuzziness makes it difficult to build reliable systems.
Generating consistent snapshots should be possible since Wikidata’s
production database contains the edit history; the generator could simply
ignore any changes to the live database that are more recent than
the snapshot time.

2. It would be nice if the dump could also include redirects, and indicate
which items have been deleted. For consistency, this should be snapshotted
at the same point in time as the actual data.

3. Statements should be sorted by subject entity ID. This would
allow data consumers to build their own data structures (eg. an LMDB
B-tree or similar) without having to re-shuffle all of Wikidata.

For this experiment, I have not bothered with any of this since it does
not affect the format. (Actually, sorting as in #3 might slightly
change the file size, perhaps making it smaller by a small amount;
but the difference is unlikely to be significant). I’m just noting this
as a wishlist for re-implementing Wikidata dumps.


## Footnotes

1. Size
    * `wikidata-20230424-all.json.br`:   72848722597 bytes =  67.8 GiB
    * `wikidata-20230424-all.json.bz2`:  81539742715 bytes =  75.9 GiB
    * `wikidata-20230424-all.json.gz`:  123717867013 bytes = 115.2 GiB
    * `wikidata-20230424-all.json.zst`:  77593744874 bytes =  72.3 GiB
    * `wikidata-20230424-all.qs.br`:     27787010556 bytes =  25.9 GiB
    * `wikidata-20230424-all.qs.bz2`:    28229997539 bytes =  26.3 GiB
    * `wikidata-20230424-all.qs.gz`:     41820873140 bytes =  38.9 GiB
    * `wikidata-20230424-all.qs.zst`:    28567267401 bytes =  26.6 GiB

2. Decompression time measured on [Hetzner Cloud](https://www.hetzner.com/cloud), Falkenstein data center, virtual machine model CAX41, Ampere ARM64 CPU, 16 cores, 32 GB RAM, Debian GNU/Linux 11 (bullseye), Kernel 5.10.0-21-arm64, data files located on a mounted ext4 volume

    * `time brotli -cd wikidata-20230424-all.json.br >/dev/null`, brotli version 1.0.9 → real 25m1.450s, user 20m6.214s, sys 0m47.106s
    * `time pbzip2 -cd wikidata-20230424-all.json.bz2 >/dev/null`, parallel pbzip2 version 1.1.13 → real 926m39.401s, user 930m39.828s, sys 3m30.333s
    * `time lbzcat -cd wikidata-20230424-all.json.bz2 >/dev/null`, lbzip2 version 2.5 → real 59m30.694s, user 943m48.935s, sys 7m30.243s
    * `time gzip -cd wikidata-20230424-all.json.gz >/dev/null`, gzip version 1.10 → real 88m44.009s, user 86m25.866s, sys 1m18.897s
    * `time zstdcat wikidata-20230424-all.json.zst >/dev/null`, zstd version 1.4.8 → real 21m46.846s, user 18m53.957s, sys 1m6.578s
    * `time brotli -cd wikidata-20230424-all.qs.br >/dev/null`, brotli version 1.0.9 → real 10m31.041s, user 8m25.385s, sys 0m17.338s
    * `time pbzip2 -cd wikidata-20230424-all.qs.bz2 >/dev/null`, parallel pbzip2 version 1.1.13 → real 93m36.174s, user 94m50.751s, sys 0m40.565s
    * `time lbzcat -cd wikidata-20230424-all.qs.bz2 >/dev/null`, lbzip2 version 2.5 → real 7m3.783s, user 109m57.272s, sys 2m19.303s
    * `time gzip -cd wikidata-20230424-all.qs.gz >/dev/null`, gzip version 1.10 → real 22m48.054s, user 22m8.762s, sys 0m21.047s
    * `time zstdcat wikidata-20230424-all.qs.zst >/dev/null`, zstd version 1.4.8 → run 1: real 5m58.011s, user 5m51.994s, sys 0m5.996s;
    run 2: real 5m55.021s, user 5m47.642s, sys 0m7.364s;
    run 3: real 5m53.228s, user 5m47.401s, sys 0m5.820s;
    average: real 5m55.420s, user 5m49.012s, sys 0m6.393s

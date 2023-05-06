# Experiment: New Format for Wikidata Dumps?

This is an experiment for a new, more compact and much faster to process
data format for [Wikidata](https://wikidata.org).

| Format      |   Size¹ |  Decompression time² |
|-------------|---------|----------------------|
| `.json.bz2` |    100% |                 100% |
| `.qs.zst`   |     35% |                 TODO |


The proposed new format takes only a third of the current size,
and decompression would be TODO times faster than today.


## Motivation

As of May 2023, the most compact format for [Wikidata
dumps](https://dumps.wikimedia.org/wikidatawiki/entities/20230424/) is
JSON with bzip2 compression.  However, the current JSON syntax is very
verbose, which makes it slow to process. Another issue is bzip2: since
its invention 27 years ago, newer algorithms have been designed that
can be decompressed much faster on today’s machines.

As a frequent user of Wikidata dumps, I got annoyed by the cost of
processing the current format, and wondered how a better format could
look. Specifically, a new format should: be significantly smaller in
size; be significantly faster to decompress; be easy to understand;
feel familiar to experienced Wikidata editors; and be easily
processable with standard libraries.

Wikidata editors frequently use the [QuickStatements
tool](https://www.wikidata.org/wiki/Help:QuickStatements) for bulk
maintenance. The tool accepts statements in a text syntax that is easy
to understand and quite compact. I wondered if Wikidata dumps could be
encoded as QuickStatements, and compressed with a modern algorithm
such as [Zstandard](https://en.wikipedia.org/wiki/Zstd).


## Extensions to QuickStatements syntax

Note that the current QuickStatements syntax cannot yet express all of
Wikidata.  The only major missing piece is ranking. For this experiment, I
used ↑ and ↓ arrows to encode preferred and deprecated rank, as in
`Q12|P9|↑"foo"`. The other missing parts are minor and rare, such as
coordinates on Venus and Mars; I did not bother for this experiment. To
express Wikidata dumps in QuickStatements format, suitable syntax
would need to be defined and properly documented. Of course, it would
then also make sense to extend the live QuickStatments tool, so it supports
the exact same syntax as the dumps.

Currently, QuickStatements does not really define an escape mechanism
for quote characters. In my experiment, I used an Unicode escape sequence
when a quoted string contained a quote, as in `"foo \u0022 bar"`.


## Other issues with Wikidata dumps

In a new version of Wikidata dumps, I think it would be good to
address some other things. From most to least important:

1. Wikidata dumps should be atomic snapshots, taken at a single point
in time. Consumers should be able to apply real-time updates to the
dump, asking for all changes since the timestamp of the dump (or
possibly the edit sequence number). Currently, this is very fuzzy,
which makes it difficult to build reliable systems based on Wikidata
dumps.

2. Statements should be sorted by entity ID. For the dump generation
process, the added expense would be marginal, and it would allow data
consumers to build their own data structures (eg. an LMDB B-tree or
similar) without having to re-shuffle all of Wikidata.

3. Better hosting. The Wikimedia dumps are currently hosted on servers
with abysmal bandwidth.  Even from within Wikimedia’s own datacenters
(specifically, Tooforge and Cloud VPS), bandwidth seems to be
throttled at about 5 MBit/s. For 2023, this is incredibly slow.

For this experiment, I have not bothered with any of this since it does
not affect the format discussion. (The sorting of #2 might slightly
change the file size, perhaps making it smaller by a small amount,
but the difference is unlikely to be significant). I’m just noting this
as a wishlist for re-implementing Wikidata dumps.


## Footnotes

1. Size
    * `wikidata-20230424-all.json.bz2`: 81539742715 bytes = 75.9 GiB
	* `wikidata-20230424-all.qs.zst`: 28567267401 bytes = 26.6 GiB
2. Decompression time measured on [Hetzner Cloud](https://www.hetzner.com/cloud), virtual machine model CAX41, 16 virtual Ampere ARM64 CPU cores, 32 GB RAM, Debian GNU/Linux 11 (bullseye), Kernel 5.10.0-21-arm64, data files located on a mounted 120 GiB volume
    * `time pbzip2 -dc wikidata-20230424-all.json.bz2 >/dev/null`, parallel pbzip2 version 1.1.13, three runs [TODO, TODO, TODO], average decompression time = TODO seconds
    * `time zstdcat wikidata-20230424-all.qs.zst >/dev/null`, zstdcat version 1.4.8, three runs [369 s, TODO, TODO], average decompression time = TODO seconds

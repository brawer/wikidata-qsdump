# Experiment: New Format for Wikidata Dumps?

This is an experiment for a simpler, smaller and faster decompressible
data format for [Wikidata dumps](https://www.wikidata.org/wiki/Wikidata:Database_download).

| Format      |     Size¹ |  Decompression time² |
|-------------|-----------|----------------------|
| `.json.bz2` |  75.9 GiB |                 TODO |
| `.qs.zst`   |  26.6 GiB |                 TODO |


The proposed format,
[QuickStatements](https://www.wikidata.org/wiki/Help:QuickStatements)
with [Zstandard](https://en.wikipedia.org/wiki/Zstd) compression, would
take about a third of the current file size. On a
typical modern cloud server, decompression would be TODO times faster
than today.


## Motivation

As of May 2023, the most compact format for [Wikidata
dumps](https://dumps.wikimedia.org/wikidatawiki/entities/20230424/) is
JSON with bzip2 compression.  However, the current JSON syntax is very
verbose, which makes it slow to process. Another issue is bzip2: since
its invention 27 years ago, newer algorithms have been designed that
can be decompressed much faster on today’s machines.

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
in time. Consumers should be able to take a dump and then apply all
changes made after its snapshot time. Currently, the snapshot time varies
across items, which makes it difficult to build reliable systems.
Generating consistent snapshots should be possible since Wikidata’s
production database contains the edit history; the generator could simply
ignore any changes to the live database that are more recent than
the snapshot time.

2. It would be nice if the dump also included redirects and the information
which items have been deleted. This should be atomically snapshotted at
the same point in time like all other data.

3. Statements should be sorted by subject entity ID. This would
allow data consumers to build their own data structures (eg. an LMDB
B-tree or similar) without having to re-shuffle all of Wikidata.

4. Better hosting. Currently, access to dump files seems to get
throttled at 5 MBit/s, even when reading from Wikimedia’s own datacenters
(Tooforge and Cloud VPS). In comparison, cheap cloud providers like
Hetzner or DigitalOcean can sequentially read data from mounted volumes
at TODO MBit/s. It’s not obvious why Wikimedia’s cloud is that much slower;
in all likelihood, Hetzner and DigitalOcean use [Ceph](https://en.wikipedia.org/wiki/Ceph_(software)) for storage, just like everyone else. Maybe it
is a hardware problem, or (more likely) at configuration problem. In any
case, it would be really nice if this could be improved.

For this experiment, I have not bothered with any of this since it does
not affect the format. (Actually, sorting as in #3 might slightly
change the file size, perhaps making it smaller by a small amount;
but the difference is unlikely to be significant). I’m just noting this
as a wishlist for re-implementing Wikidata dumps.


## Footnotes

1. Size
    * `wikidata-20230424-all.json.bz2`: 81539742715 bytes = 75.9 GiB
	* `wikidata-20230424-all.qs.zst`: 28567267401 bytes = 26.6 GiB
2. Decompression time measured on [Hetzner Cloud](https://www.hetzner.com/cloud), Falkenstein data center, virtual machine model CAX41, Ampere ARM64 CPU, 16 cores, 32 GB RAM, Debian GNU/Linux 11 (bullseye), Kernel 5.10.0-21-arm64, data files located on a mounted 120 GiB volume
    * `time pbzip2 -dc wikidata-20230424-all.json.bz2 >/dev/null`, parallel pbzip2 version 1.1.13, TODO user/real/system secondcs
    * `time zstdcat wikidata-20230424-all.qs.zst >/dev/null`, zstdcat version 1.4.8, three runs [369 s, TODO, TODO], average decompression time = TODO seconds

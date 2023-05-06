# Wikidata QuickStatements Dump

This is an experiment for a new, more compact and faster to process
data format for the [Wikidata](https://wikidata.org) knowledge graph.

Currently, Wikidata gets dumped into a large JSON file that gets compressed
with the bzip2 compression algorithm. However, the current JSON syntax is
very verbose, which makes it slow (and memory-intensive) to parse. Also,
decoding a bzip2-compressed stream is very CPU-intensive. In combination,
this makes it expensive (in terms of CPU, memory and ultimately money)
to process Wikidata. The purpose of this experiment is to quantify the
gains if Wikidata were to adopt a more compact format and a more modern
compression algorithm. Specifically, this experiment converts the current
Wikidata dump to [QuickStatements](https://www.wikidata.org/wiki/Help:QuickStatements) format (with small extensions, eg. to express deprecated and preferred
rank) in [Zstandard](https://en.wikipedia.org/wiki/Zstd) compression.


| Format      |   Size¹ |  Decompression time² |
|-------------|---------|----------------------|
| `.json.bz2` |    100% |                 100% |
| `.qs.zst`   |     35% |                 TODO |


1. Size
    * `wikidata-20230424-all.json.bz2`: 81539742715 bytes = 75.9 GiB
	* `wikidata-20230424-all.qs.zst`: 28567267401 bytes = 26.6 GiB
2. Decompression time measured on [Hetzner Cloud](https://www.hetzner.com/cloud), virtual machine model CAX41, 16 virtual Ampere ARM64 CPU cores, 32 GB RAM, Debian GNU/Linux 11 (bullseye), Kernel 5.10.0-21-arm64
    * `time pbzip2 -dc wikidata-20230424-all.json.bz2 >/dev/null`, parallel pbzip2 version 1.1.13, three runs [TODO, TODO, TODO], average decompression time = TODO seconds
    * `time zstdcat wikidata-20230424-all.qs.zst >/dev/null`, zstdcat version 1.4.8, three runs [369 s, TODO, TODO], average decompression time = TODO seconds





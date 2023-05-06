# Experiment: New Format for Wikidata Dumps?

This is an experiment for a new, more compact and much faster to process
data format for [Wikidata](https://wikidata.org).

| Format      |   Size¹ |  Decompression time² |
|-------------|---------|----------------------|
| `.json.bz2` |    100% |                 100% |
| `.qs.zst`   |     35% |                 TODO |


As of May 2023, the most compact format for [Wikidata dumps](https://dumps.wikimedia.org/wikidatawiki/entities/20230424/) is JSON with bzip2 compression.
(Additionally, there’s also other encodings, which we ignore for this
discussion because they taken even more space than `.json.bz2`).
we ignore them because they are even more verbose).
However, the current JSON syntax is quite verbose, which makes it slow
to process. Also, the bzip2 compression compression algorithm has been
designed almost 30 years ago; meanwhile, newer algorithms have been designed
that can be decompressed much faster on today’s machines.

As a frequent user of Wikidata dumps, I got annoyed by the cost of
processing the current format, and wondered how a better format
could look. Specifically, a new format should:

* be significantly smaller in size;
* be significantly faster to decompress;
* be easy to understand;
* feel familiar to experienced Wikidata editors;
* be easily processable with standard libraries.

For bulk maintenance, Wikidata editors typically use the
[QuickStatements tool](https://www.wikidata.org/wiki/Help:QuickStatements).
The tool takes editing commands in a text syntax that is easy to understand,
very compact, and it is already familiar by many Wikidata maintainers.

Therefore, I wondered whether Wikidata dumps could be encoded
as QuickStatements with a modern compression algorithm such as
[Zstandard](https://en.wikipedia.org/wiki/Zstd). Note
that the current QuickStatements syntax does not (yet) cover the entire
Wikidata semantics. The biggest missing part is being able to
express preferred and deprecated ranks. For the purpose of this experiment,
I came up with an ad-hoc encoding that uses ↑ and ↓ arrows for  rank, as in
`Q123|P987|↑"foo bar"@en`. The other missing parts are super minor,
such as coordinates of places that are located on other planets than Earth.
Since there is very little such data, I decided to skip such statements
from this experiment. Of course, these should have to be implemented
(by defining a QuickStatements syntax, and supporting it in the live
QuickStatments tool) if Wikidata wanted to adopt QuickStatements as
its new dump format.


1. Size
    * `wikidata-20230424-all.json.bz2`: 81539742715 bytes = 75.9 GiB
	* `wikidata-20230424-all.qs.zst`: 28567267401 bytes = 26.6 GiB
2. Decompression time measured on [Hetzner Cloud](https://www.hetzner.com/cloud), virtual machine model CAX41, 16 virtual Ampere ARM64 CPU cores, 32 GB RAM, Debian GNU/Linux 11 (bullseye), Kernel 5.10.0-21-arm64
    * `time pbzip2 -dc wikidata-20230424-all.json.bz2 >/dev/null`, parallel pbzip2 version 1.1.13, three runs [TODO, TODO, TODO], average decompression time = TODO seconds
    * `time zstdcat wikidata-20230424-all.qs.zst >/dev/null`, zstdcat version 1.4.8, three runs [369 s, TODO, TODO], average decompression time = TODO seconds





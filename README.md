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





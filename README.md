# tendigitprimes
A complete list of prime numbers up to 9999999999, split into 5M line files, and web-service-ready

___________


# Motivation

Prime numbers are used in a bunch of scenarios due to their uniqueness, being numbers that are not divisible by another 
number. Many use cases range from mathematics to cryptography, and my original interest was to retrieve secure constants 
for obfuscation logic.

The need for this type of logic can be to provide a public reference to an internal identifier, so an organization is 
able to supply this reference without ever giving out any private information by obfuscating it. Then, the organization 
is able to consume the same obfuscated identifier and extract the internal identifier it represents.  

Data obfuscation is not the target in this repository. But the logic contained a 10-digit prime number as a constant; 
that allows it to encode and decode data on-the-fly. So, it would technically be possible (with the right bounds and the
right math) to, one day, create an obfuscator service that provides data encoders and decoders according to a 
specification (or, prime number). `tendigitprimes` would be the provisioning service that supplies 10-digit primes 
on-demand, to a consumer. 

# Growth

This repository started as a bunch of text files with 5-million numbers separated by newlines, as this was (at the time)
the most portable way I could publish these numbers publicly for ease-of-access. In 2022, I wasn't yet focused in 
creating a service for querying these numbers; but to store it somewhere besides my computer. This data can be useful 
to a lot of use-cases, so I made it public since the beginning.

As time went by and I explored Go more and more, as well as other technologies, I had the chance to revisit this 
repository. This time, the intention was to create a simple web service that I could use to query for a random prime 
number (optionally, with some limits).

The go-to technology stack was simple: SQLite with gRPC (and gRPC-gateway for _regular_ HTTP calls). This is a simple 
and usual stack for such a service that needs to keep state (to query these primes from the database), yet it will be 
read-only. The problem was that the queries were taking 3, 4 minutes to complete. This brings us to the challenges faced.

____________

# Challenges

The main problem is performing a `SELECT` query with a comparison (either `WHERE ... AND ...` or `BETWEEN ...`) against
so many items.

If we count the number of primes in `./raw`, it's close to half a billion numbers:

```shell
❯ wc -l ./raw/*
# ...
 455052511 total
```

A few ideas that I explored were around the table structure, such as having or not having a primary key, as well as 
creating indices or not. The latter was not helpful in this case due to the narrow "shape" of the data -- it's simply 
one column with an integer value, and SQLite would always create an index for each row under-the-hood. Creating indices 
would only create bigger and bigger database files with no gain (maybe loss) in performance.

This was very disappointing for someone so into SQLite as me.

So, I tried PostgreSQL and MongoDB, and didn't find an immediate gain. With Postgres I could conclude that the insert 
queries taking such a long time was not a SQLite problem -- it was really just a matter of data volume. With Mongo, I 
just dropped it (just like Postgres) since I really wanted a portable, standalone solution.

Regarding the data insert step, it was possible to speed it up greatly by adding the data in a SQL transaction,
multiple values at a time (32000 to be precise). This reduces the overhead of starting a statement and executing it.

However, the queries were still taking too long; the best results I was getting with indices and all the `PRAGMA` 
optimizations I could find was around a 1:30 minute query, to get a prime number from 1000000000 to 5000000000. Even 
trying to load all data into an in-memory instance and querying it was showing the same performance and results.

They (the internet) told me we shouldn't do this but I did it anyway, and _this_ is partitioning SQLite databases. While
I understand that SQLite really shines on local-data management and storage, I still saw it as a great go-to candidate
for this pursuit.

It turns out that huge tables are super impactful: 

- The first attempts at generating a whole SQLite database file with all the numbers took approximately double the space 
that the `./raw` directory took. At first this was due to an addition of an `id` primary key field in the table schema. 
The primary key was dropped and the `primes` table only contains a `prime` element, that is also primary key.
- Adding any indices to the table resulted in exponentially larger database files (14GB, 21GB). I tried to add indices 
that theoretically would make sense when querying values with limits, putting sets of numbers in buckets. This didn't 
actually provide any performance benefits, likely due to the added size in the database file itself.
- With this in mind, it did occur to me to try creating different tables -- but it was also clear that the database file
size was being impactful; so I needed a better solution to distribute the load.

Hitting the wall on these key-points allowed me to finally give it a shot and partition the data into different 
SQLite database files.

Creating the database files is very straight-forward and the logic mostly in Go, we just need to parse the numbers in 
the `./raw` directory files, and creating several primes databases for all of those numbers based on simple min-max 
rules.

The direction for creating the partitions was to split the numbers in buckets representing ranges (from 0 to 499999999, 
from 500000000 to 999999999, etc.), and saved onto different files (partitions). Alongside this process, there is also 
an index database keeping track of:
- each partition's ID (that will be in its filename)
- each partition's minimum and maximum values, for its range
- each partition's total amount of prime numbers

So far, so good. And regarding disk space, surprise-surprise! It was now taking **less space** than the raw files!

```shell
❯ du -h ./sqlite/parts                                                            
4.3G    ./sqlite/parts

❯ du -h ./raw                                                                        
4.6G    ./raw
```

To orchestrate operations against all these databases (a total of 100 partitions), the easiest option would be 
using SQLite's `ATTACH DATABASE` statement. This was easy as pie since you can reference the database in a query, in the 
table clause as a schema (preceding it with a `.`).

The only problem is that the default SQLite limit for attached databases is `10`, and I had another 90 databases to 
attach. Back to research. The maximum amount of databases allowed by SQLite is actually 125 databases but these limits 
can only be raised using compile-time options (the runtime options for setting this limit, e.g. via `PRAGMA`s only 
change soft limits -- or, lowering the maximum for that specific connection).

I honestly learned a lot on how SQLite works, particularly how the `modernc.org/sqlite` library is doing it with a 
CGO-free implementation. I got very close to recompiling the `modernc.org/libsqlite3` repository using my custom value 
of `100` for the maximum attached databases; however due to some C-compiler and machine-related issues I was running 
into errors over and over again. For a quick-fix, I just cloned the `sqlite` repository and modified the constants 
within the `lib` package (as generated from `libsqlite3`) related to the maximum of attached databases. From here, I 
just needed to setup a Go workspace using my version of the library and the app was then ready to attach 100 databases!

To wrap things up, I added a new implementation for the repository interface that the `primes` package consumes, that 
handles the partitioned databases. The initial difference (just querying for numbers on different partitions, with a 
`WHERE ... BETWEEN ...` clause) was already mind-blowing, from the until-then best time of 1 minute and 30 seconds to 
1 and a half second (60x improvement only by partitioning).

From this point, purely for the sake of going below the 1-second mark, I tried:
- using a "random index" approach to get _any prime number_ within a bucket; using `LIMIT 1 OFFSET {num}`, where `num` 
is a random number within the total amount of primes in this partition.
- After the query, the number is evaluated (according to the input filter) if it is within the request's bounds.
- Just keep fetching random primes (within certain bounds) to either return a random prime (one iteration) or to list 
primes (a set of random primes within bounds) instead of listing a bunch of them and returning a sequential set or 
picking random ones.
- Randomly selecting different partitions in a set of possible ones, by adding a random number between zero and the 
total amount of partitions to an index value, that is bound by a modulo operation over the total amount of partitions.

Using these techniques, it was possible to shave time off from 1 second and a half to about 150ms per random prime 
request, and takes about 1 second and a half to list 50 primes within the same bounds 
(`?min=1000000000&max=5000000000&max_results=50`). Also, important to note, these times include the HTTP overhead from 
gRPC and gRPC-gateway, as the measures are from `curl` calls against a `localhost` runtime of this service.

Awesome ride with SQLite, and allowed me to also contribute to the `modernc.org/sqlite` repository.

____________

# Usage

> You can take a closer look into the scripts directory for database builder scripts, but this guide will
> focus on creating a partitioned SQLite instance just like the final version I described in [Growth](#growth)

## Generate SQLite databases from the raw data

It all begins with the text files under the `raw` directory. They were the initial target format when generating the 
entire set, and they are still the reference for creating the SQLite databases.

Locally, you can create a target directory (say, `parts`) that you will define as an output **directory** for your database 
index and partitions. This is through the `-o` (output) flag.

Also, since this guide involves partitioning, you will also want to add the `-p` (partitioned) flag:

```shell
go run ./scripts/build-sqlite -p -o ~/path/to/my/parts 
```

This operation takes a good amount of time (about 1 hour and a half on an M1 Pro Mac), so the log output is verbose 
enough to provide context on the progress. If the execution fails or is halted, please remove any generated files within 
your output directory and generate them again.  

## Custom build SQLite

For this amount of partitions to work, a custom build of SQLite is required. Since this app uses `modernc.org/sqlite` as 
a SQL driver, then it's advisable to follow their guides around rebuilding / regenerating the library files. This is 
done by cloning `modernc.org/libsqlite3` and running `make`, however you should always look into specifics that could 
change over time.

When generating the library, you will want to impose a new limit for `SQLITE_MAX_ATTACHED`. This is done with a 
`go generate` call setting the sqlite configuration variables in the `GO_GENERATE` env var:

```shell
GO_GENERATE=-DSQLITE_MAX_ATTACHED=100 go generate
```

An alternative to recompiling SQLite is, like I described (as a hack!):
- cloning the `modernc.org/sqlite` library
- find all occurrences within the `lib` directory for `const SQLITE_MAX_ATTACHED = 10`. This is a Go constant that is 
modified with the `go generate` step described above. You can find and replace this line with a new limit, like:

```
-const SQLITE_MAX_ATTACHED = 10
+const SQLITE_MAX_ATTACHED = 100
```

> this change should at least apply to the architecture of the machine that will run the service, although you can 
> simply apply the change to all target architecture files.

- initialize a Go workspace that includes your modified copy of `sqlite` and this (`tendigitsprimes`) repository, e.g.:

```shell
go work init ./ ../../../gitlab.com/cznic/sqlite 
```

## Running the service

With all the above prepared, you should be able to serve the primes without any issues. The command below sets two 
environment variables, the first one points to the databases' directory as when generated, and the second indicates that
we are using a partitioned repository.

```shell
PRIMES_DB_URI=~/path/to/my/parts PRIMES_DB_IS_PARTITIONED=1 go run ./cmd/primes serve
```

## Using the service

You can issue HTTP GET requests to both fetch a random prime or a list of them:

### Random

This RPC returns a random prime number up to 10 digits in length:

```http request
GET /v1/primes/rand?min=1000000000&max=5000000000
Host: localhost:8080
Content-Type: application/json

{}
```

Example response: 

```json
{
  "prime_number": "4696898233"
}
```

### List

This RPC returns a set of prime numbers up to 10 digits in length:

```http request
GET /v1/primes?min=1000000000&max=5000000000&max_results=10
Host: localhost:8080
Content-Type: application/json

{}
```

Example response:

```json
{
  "prime_numbers": [
    "1032472541",
    "1125602999",
    "2948325041",
    "4213413031",
    "4598902999",
    "4552085387",
    "3962748233",
    "1651015459",
    "4819666693",
    "4412108893"
  ]
}
```
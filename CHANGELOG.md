# Changelog

## [v0.1.17](https://git.alanpearce.eu/searchix/diff/?id=v0.1.16&id2=fb0ab22c9a8ca1784477e64b03d5cb62d607bc9f) (2025-01-19)

### Fixes

-   **sentry:** report correct HTTP path in traces
    ([eece88d](https://git.alanpearce.eu/searchix/commit/?id=eece88d32b1b613aba7ec2f11bcfeb10163616b1))
-   use importer configured timeout for sentry check-ins
    ([b706b5a](https://git.alanpearce.eu/searchix/commit/?id=b706b5a3d09983637ea6d5af515733d77b674bdb))

### [v0.1.16](https://git.alanpearce.eu/searchix/diff/?id=v0.1.15&id2=v0.1.16) (2025-01-15)

#### Features

-   enable searching via program names for multi-program packages
    ([7aea6aa](https://git.alanpearce.eu/searchix/commit/?id=7aea6aa210a8939ac208fb7540d1b46ba69a995f))

#### Fixes

-   **dev:** fix CSP not working in dev mode behind reverse proxy
    ([b26ddba](https://git.alanpearce.eu/searchix/commit/?id=b26ddba432f8bde78022d2fc8837f0ffb25448b1))

### [v0.1.15](https://git.alanpearce.eu/searchix/diff/?id=v0.1.14&id2=v0.1.15) (2025-01-05)

#### Features

-   offer to show all results on one page (if < 10000)
    ([9207642](https://git.alanpearce.eu/searchix/commit/?id=920764211064521930441b96ed914034302a7470))

#### Fixes

-   **dev:** overwrite previous asset hash on change
    ([d289e54](https://git.alanpearce.eu/searchix/commit/?id=d289e54af28de6b21cbe49cebc262c4e3ca9f6d2))

### [v0.1.14](https://git.alanpearce.eu/searchix/diff/?id=v0.1.13&id2=v0.1.14) (2025-01-02)

#### Features

-   show scores in results table in dev mode
    ([58643f3](https://git.alanpearce.eu/searchix/commit/?id=58643f38bdd62dc538a03865b1f81eb77d957779))

#### Fixes

-   boost non-stemmed matches over stemmed matches
    ([797c6e2](https://git.alanpearce.eu/searchix/commit/?id=797c6e27e18f6c47ceb4266d747a05064365c32d))

### [v0.1.13](https://git.alanpearce.eu/searchix/diff/?id=v0.1.12&id2=v0.1.13) (2024-12-30)

#### Features

-   increase results per page
    ([e1787bf](https://git.alanpearce.eu/searchix/commit/?id=e1787bf0b6dbac3a1601559d2f00ebc4b2df0f29))
-   de-rank packages on 'all' search
    ([df5db96](https://git.alanpearce.eu/searchix/commit/?id=df5db969262db1382dcd9310d362efdf8f8b6949))

### [v0.1.12](https://git.alanpearce.eu/searchix/diff/?id=v0.1.11&id2=v0.1.12) (2024-12-04)

#### Fixes

-   use UTC time for fetch/import timer
    ([c92530b](https://git.alanpearce.eu/searchix/commit/?id=c92530b7940bfd9e0940dd07e4a33b8dc4b575ea))
-   enable using module in pure evaluation mode
    ([ec4946e](https://git.alanpearce.eu/searchix/commit/?id=ec4946ee959b2d7d28287e9cd4643a0698833f6b))
-   enable using module in pure evaluation mode
    ([174ad04](https://git.alanpearce.eu/searchix/commit/?id=174ad04905da9709e4aef366e14aa438e58e5b0b))

### [v0.1.11](https://git.alanpearce.eu/searchix/diff/?id=v0.1.10&id2=v0.1.11) (2024-11-20)

#### Features

-   apply stemming to search index terms
    ([72d56a6](https://git.alanpearce.eu/searchix/commit/?id=72d56a6781e97cde13ba068e8553e9342bd1fa20))

### [v0.1.10](https://git.alanpearce.eu/searchix/diff/?id=v0.1.9&id2=v0.1.10) (2024-11-10)

#### Fixes

-   search all missing expected results
    ([8656f47](https://git.alanpearce.eu/searchix/commit/?id=8656f47cd62b084a59546c79d437a120869661e3))

### [v0.1.9](https://git.alanpearce.eu/searchix/diff/?id=v0.1.8&id2=v0.1.9) (2024-08-17)

#### Fixes

-   don't error when item not found
    ([b427c26](https://git.alanpearce.eu/searchix/commit/?id=b427c2657582caa9b2cce8bcfd82624663dda780))
-   tell browsers not to capitalise search input
    ([faf4d0e](https://git.alanpearce.eu/searchix/commit/?id=faf4d0e5f21266bdad99c991efef4e83c09da5c6))

### [v0.1.8](https://git.alanpearce.eu/searchix/diff/?id=v0.1.7&id2=v0.1.8) (2024-07-13)

#### Features

-   enable combined search
    ([6781684](https://git.alanpearce.eu/searchix/commit/?id=6781684647ad7f131054a93e1d13105d865c0e83))
-   more structured logging
    ([94b21b2](https://git.alanpearce.eu/searchix/commit/?id=94b21b286edff37496a2fe481963625ac01c30a1))

#### Fixes

-   nil panics rendering last page of search results
    ([b1fd710](https://git.alanpearce.eu/searchix/commit/?id=b1fd7104004b9646f541afbc0add3047dbff59a7))
-   panic rendering error pages
    ([ee83dfd](https://git.alanpearce.eu/searchix/commit/?id=ee83dfd741c47bd25eac1ac7139f20875de0452c))
-   opensearch searches not using input query
    ([a5f58bc](https://git.alanpearce.eu/searchix/commit/?id=a5f58bc7fe205b8ff84f6b654513e40bfbc61ae9))
-   extra data after error parsing query string
    ([4423458](https://git.alanpearce.eu/searchix/commit/?id=442345898802e93b0b884fd0f1c3492e08c34049))
-   panic rendering search results
    ([5a2d54d](https://git.alanpearce.eu/searchix/commit/?id=5a2d54d9c8751935284293b2ffd396fff0a5d202))
-   panic on search error
    ([f076b5b](https://git.alanpearce.eu/searchix/commit/?id=f076b5bd6cb82edd99be50f3dbdd39bb9be2c44e))

### [v0.1.7](https://git.alanpearce.eu/searchix/diff/?id=v0.1.6&id2=v0.1.7) (2024-06-29)

#### Features

-   add message for users with javascript disabled
    ([e500a10](https://git.alanpearce.eu/searchix/commit/?id=e500a1089797eee1b027d5237b2e4f140004098c))

#### Fixes

-   blank modal when using browser navigation
    ([8ef4bc9](https://git.alanpearce.eu/searchix/commit/?id=8ef4bc9b1f16dd6395d0e0daa85ada9b8db12015))
-   duplicated results tables via JS search
    ([e0d86be](https://git.alanpearce.eu/searchix/commit/?id=e0d86bec7c9e8bedd1e71173b486668e16ec7c2e))
-   panic rendering results
    ([dbeddce](https://git.alanpearce.eu/searchix/commit/?id=dbeddced4046cb5756af2604f5cb6734b3e7cee6))
-   incorrect formatting of options results headers
    ([107d54a](https://git.alanpearce.eu/searchix/commit/?id=107d54a4c4a415f3581a452878cdd625f7f2634a))

### [v0.1.6](https://git.alanpearce.eu/searchix/diff/?id=v0.1.5&id2=v0.1.6) (2024-06-19)

#### Features

-   add low-memory mode
    ([63599c7](https://git.alanpearce.eu/searchix/commit/?id=63599c719d8e4e09548546bb366a849ef40c602b))
-   enable configuration of flags as environment variables
    ([a495a67](https://git.alanpearce.eu/searchix/commit/?id=a495a673d2dda7b2aa9b11217d4662367a627517))
-   display results in a table, showing details on click
    ([408aed0](https://git.alanpearce.eu/searchix/commit/?id=408aed03d3454330120475ca53838a2f4fe28ea3))
-   add detail pages for packages/options
    ([d40c0e1](https://git.alanpearce.eu/searchix/commit/?id=d40c0e188a7fe1b36887f59c4a9958faa81b3d44))
-   use stale-while-revalidate instead of immutable asset paths
    ([872ae5d](https://git.alanpearce.eu/searchix/commit/?id=872ae5db32c605ef385588a33673ce5fef29ff21))
-   serve assets via immutable paths
    ([896926a](https://git.alanpearce.eu/searchix/commit/?id=896926a63a8f2d145467b325f9b0198315e0af6d))
-   make security headers stricter
    ([b537694](https://git.alanpearce.eu/searchix/commit/?id=b53769462bf830f860b7d741a3d0801afdbc9aa2))
-   enable sub-resource integrity for assets
    ([4698a97](https://git.alanpearce.eu/searchix/commit/?id=4698a97974ae82e7bd8592828c58294b222a58ff))

#### Fixes

-   create index metadata when index exists and meta doesn't
    ([e665146](https://git.alanpearce.eu/searchix/commit/?id=e6651469b9e934cb4dc65f760277111a79e9ae2d))
-   moving through search history shows no results
    ([66b2556](https://git.alanpearce.eu/searchix/commit/?id=66b2556a6a7c911a69b231fddeefe0a939d8898d))
-   nil pointer exception when (re-)indexing
    ([2c4a19a](https://git.alanpearce.eu/searchix/commit/?id=2c4a19ad6cb71065d31304370708c859b4f68f6d))

### [v0.1.5](https://git.alanpearce.eu/searchix/diff/?id=v0.1.4&id2=v0.1.5) (2024-05-29)

#### Features

-   improve search relevance by boosting keyword matches
    ([73185e4](https://git.alanpearce.eu/searchix/commit/?id=73185e4633d4f56ce8acdd3cc19f78faaa78bd2b))
-   support cleartext HTTP/2 (i.e. from a reverse proxy)
    ([16b4337](https://git.alanpearce.eu/searchix/commit/?id=16b433791b619f4144110e4429d98684bf5e98eb))

#### Fixes

-   missing alternate forms of packages with the same name
    ([838e178](https://git.alanpearce.eu/searchix/commit/?id=838e178e503d9576a40fce8fc986c4e1972e4c0f))
-   incorrect (switched) behaviour of --update/--replace
    ([97021f7](https://git.alanpearce.eu/searchix/commit/?id=97021f769a0f06103b4af5019d942759d644e291))
-   re-enable exit --update/--replace
    ([a5e3c14](https://git.alanpearce.eu/searchix/commit/?id=a5e3c144034eab0cb41c4b65557e5c30e877c9fb))
-   check for nil pointers in NewMux
    ([2c1491d](https://git.alanpearce.eu/searchix/commit/?id=2c1491de56d0c3e2f4cb0b0c1e33035510f72fc5))

### [v0.1.4](https://git.alanpearce.eu/searchix/diff/?id=v0.1.3&id2=v0.1.4) (2024-05-23)

#### Features

-   stream files directly from fetcher to importer
    ([0dbfe37](https://git.alanpearce.eu/searchix/commit/?id=0dbfe37fbddb95c184d845c79bbe014597d55fe8))
-   add --update flag to force-update index and exit
    ([9c790c1](https://git.alanpearce.eu/searchix/commit/?id=9c790c1c620e46ca8d2cf49c50951635621e9a7d))
-   **config:** use enums' string representations to print defaults
    ([a90419a](https://git.alanpearce.eu/searchix/commit/?id=a90419aa46ec42588f65e6327559ce479f9c5b67))
-   **config:** print durations and URLs with human values
    ([0f55080](https://git.alanpearce.eu/searchix/commit/?id=0f5508013d776f8806a87957f80e74c7f74c72eb))
-   show default configuration with --print-default-config
    ([60a1569](https://git.alanpearce.eu/searchix/commit/?id=60a15699657cef86d1a644c0d13b0d117a818633))

#### Fixes

-   abort import of source on batch processing errors
    ([3053e41](https://git.alanpearce.eu/searchix/commit/?id=3053e41b1528ef898cccd44e056e4d167619af6b))
-   **web:** fetch host from request, not header, where it's expected
    ([e69d785](https://git.alanpearce.eu/searchix/commit/?id=e69d78542a4ec92da5931b3db105b14a39a4c964))
-   **nix:** generate default baseURL from listen address and port
    ([027db43](https://git.alanpearce.eu/searchix/commit/?id=027db4372e82e445c531c958c2da0f9cf6ae1e03))
-   make fetcher check on creation that it supports Source.Importer
    ([b77a24f](https://git.alanpearce.eu/searchix/commit/?id=b77a24f9f75378ffe97be83cf4dfd7f1683b9a7e))
-   **importer:** abort import of current source in case of error
    ([c821da1](https://git.alanpearce.eu/searchix/commit/?id=c821da1cf55864852bcd8f337dd7acd0cc02f0b9))
-   nix cannot read date/time TOML values from default config
    ([202317d](https://git.alanpearce.eu/searchix/commit/?id=202317df1dba891654bf749c3720fbd602df04ff))
-   duplicate error logging on config parse error
    ([6037d5c](https://git.alanpearce.eu/searchix/commit/?id=6037d5c33e9b98360f61a4fe4f09659d0e3e6160))

### [v0.1.3](https://git.alanpearce.eu/searchix/diff/?id=v0.1.2&id2=v0.1.3) (2024-05-17)

#### Features

-   **ui:** enable resetting form by clicking current source link
    ([6ea953b](https://git.alanpearce.eu/searchix/commit/?id=6ea953b1c9f9c352a9619e42a6f0ce80d183d7f8))
-   render markdown in packages' LongDescription
    ([934a069](https://git.alanpearce.eu/searchix/commit/?id=934a0693213df901078cd1bc75a5b7e1b5c87555))
-   use attribute instead of name for packages
    ([698881a](https://git.alanpearce.eu/searchix/commit/?id=698881aa577d2c209f162db9263fab17ced3643f))

### [v0.1.2](https://git.alanpearce.eu/searchix/diff/?id=v0.1.1&id2=v0.1.2) (2024-05-17)

#### Fixes

-   **css:** pagination styles not applied
    ([f10543d](https://git.alanpearce.eu/searchix/commit/?id=f10543d0aff0e203dd57287dc2b736c686e69102))

### [v0.1.1](https://git.alanpearce.eu/searchix/diff/?id=v0.1.0&id2=v0.1.1) (2024-05-17)

#### Fixes

-   **nix:** missing version variable in build
    ([8a1c2e8](https://git.alanpearce.eu/searchix/commit/?id=8a1c2e805b70ea4c4730fe7d785b942f46a54dea))

## v0.1.0 (2024-05-17)

### Features

-   display more information about packages
    ([8df293c](https://git.alanpearce.eu/searchix/commit/?id=8df293c0ad8b8be9f43f2d5dc1334b8546d6f470))
-   error on non-matching configuration entries
    ([6488132](https://git.alanpearce.eu/searchix/commit/?id=64881323e89d6d9b96865a5fa6f2150b826683db))
-   automatically add/remove sources to/from index
    ([9f241bf](https://git.alanpearce.eu/searchix/commit/?id=9f241bff227608dd53a250d012116077dce6dab6))
-   support searching packages
    ([05e6197](https://git.alanpearce.eu/searchix/commit/?id=05e61978906a08132c4340a5f9ae518134dd0fa9))
-   **index:** enable storage of package data
    ([42611df](https://git.alanpearce.eu/searchix/commit/?id=42611df8133fc88bac5947a65a18fa095d68a951))
-   improve accessibility and browser behaviour
    ([054d2b3](https://git.alanpearce.eu/searchix/commit/?id=054d2b3bf7eb4c5a5401565e86031f22ea754062))
-   exit after rebuilding index if `--replace` is specified
    ([d6e117b](https://git.alanpearce.eu/searchix/commit/?id=d6e117bb9aeb2bae23dedd51b59b6e6f139e8a4f))
-   decode brotli-compressed files
    ([b8da487](https://git.alanpearce.eu/searchix/commit/?id=b8da487f774aa2399b79bbde5f2e6e13be4fce1c))
-   extract and store package set (e.g. nodePackages)
    ([8b78a98](https://git.alanpearce.eu/searchix/commit/?id=8b78a98af573a77a6d6aa6a0b231d4b657e35726))
-   **importer:** process packages from nixpkgs package.json
    ([51f774a](https://git.alanpearce.eu/searchix/commit/?id=51f774aaced998ad3c7c31e370aea22cc2fdbbc5))
-   add caching header (except on error pages)
    ([eb88ab1](https://git.alanpearce.eu/searchix/commit/?id=eb88ab15adaf23a61c783afe76b646855814d38a))
-   monitor import jobs via sentry
    ([fa97ca8](https://git.alanpearce.eu/searchix/commit/?id=fa97ca832d6dc301dabb5d9c76730630f99534e0))
-   add browser search engines via opensearch description
    ([5521173](https://git.alanpearce.eu/searchix/commit/?id=5521173ea0e05bda93096b51f67c5e813cb1d87e))
-   print clickable URL on startup in dev mode
    ([121ec2b](https://git.alanpearce.eu/searchix/commit/?id=121ec2b445636cd3acb59391e1b241abfe2954a5))
-   render HTML error pages
    ([2d5ee7b](https://git.alanpearce.eu/searchix/commit/?id=2d5ee7b69cbe27e7e1abef7bf3451cb6455c4387))
-   **module:** add default for Web.Environment
    ([634534d](https://git.alanpearce.eu/searchix/commit/?id=634534d3d22d8f458a70a1891d061c3cdefd4dd1))
-   add DownloadOptions importer (fetches pre-built options.json)
    ([df3b6bf](https://git.alanpearce.eu/searchix/commit/?id=df3b6bfd571c38aa7099838aa43bb3aff0eff2bf))
-   log on startup if index doesn't exist
    ([5f9c7ef](https://git.alanpearce.eu/searchix/commit/?id=5f9c7ef00b1ed56fb1abbade4d5dbc3041acd4fc))
-   rely on system to show log times in production
    ([37afa6c](https://git.alanpearce.eu/searchix/commit/?id=37afa6ce04434b81fcb242f1160938f1cddbde1e))
-   allow adding extra HTML to the body via configuration
    ([7077a37](https://git.alanpearce.eu/searchix/commit/?id=7077a3748fa545e8dee83d4f3464c55b4b459183))
-   enable using channel URLs if unable to resolve via NIX_PATH
    ([fe79b45](https://git.alanpearce.eu/searchix/commit/?id=fe79b4509dd580542fa28186673826df603677bb))
-   nix module
    ([fb25a97](https://git.alanpearce.eu/searchix/commit/?id=fb25a97b6c55c6d51e59d983fdb74b4917c97bf0))
-   clicking another source re-uses query
    ([42bcd10](https://git.alanpearce.eu/searchix/commit/?id=42bcd10f010f884bddb0142c31b526faace24186))
-   **js:** save and restore input state with history
    ([b517b1e](https://git.alanpearce.eu/searchix/commit/?id=b517b1e0e1933b310bcbbd718fc664f87226d3b6))
-   version search index mapping schema and warn if outdated
    ([76f18d6](https://git.alanpearce.eu/searchix/commit/?id=76f18d677b9b0bb70f701824e715f2ed10aeb8a0))
-   dynamically link to sources
    ([631c02a](https://git.alanpearce.eu/searchix/commit/?id=631c02ac689b3366987918737e2b09f52fb16dda))
-   format server listen address as URL
    ([f42dfe7](https://git.alanpearce.eu/searchix/commit/?id=f42dfe7b5fd3d60c9558eb520d33c7c8fd1ba201))
-   enable index path to be configured
    ([0986e86](https://git.alanpearce.eu/searchix/commit/?id=0986e8699e87addccbef43c0ce30a466988fed15))
-   **importer:** make timeout configurable via command-line
    ([1db7c2a](https://git.alanpearce.eu/searchix/commit/?id=1db7c2a1630c3d80c3c696c8fabaa5476eb2184e))
-   **importer:** add --replace flag to overwrite existing index
    ([778e1d7](https://git.alanpearce.eu/searchix/commit/?id=778e1d774f574f45c179d0fecf0c1da9cc359b80))
-   allow configuration of log level
    ([99d5507](https://git.alanpearce.eu/searchix/commit/?id=99d5507b223baa6c780707d32c7bbbb8611034f0))
-   enable setting config file path via command line
    ([14c1dda](https://git.alanpearce.eu/searchix/commit/?id=14c1ddaf7c5468c5966b0c1585f19a5ff39c206d))
-   import sources from configuration in go code and index options
    ([e062ca7](https://git.alanpearce.eu/searchix/commit/?id=e062ca72b222b890e345548bd8422d5df98e9fef))
-   make configuration optional
    ([967f6fd](https://git.alanpearce.eu/searchix/commit/?id=967f6fdf5c1693d3aa27079b3ae28768fb7356c6))
-   improve search by using a custom tokeniser for option names
    ([24b81a6](https://git.alanpearce.eu/searchix/commit/?id=24b81a63620667cd9c9f50952bbe881c08f16048))
-   paginate search results
    ([973345a](https://git.alanpearce.eu/searchix/commit/?id=973345ad50f9b237714fcb364cf7f665b3909f9d))
-   search multiple sources
    ([212e5cf](https://git.alanpearce.eu/searchix/commit/?id=212e5cf6621c99e46dbb37c860dab8938968bb19))
-   search one set of options
    ([c15b142](https://git.alanpearce.eu/searchix/commit/?id=c15b142b18dcdc7f5ab6d5f1afca8ae1696692cc))
-   link to exact commits in NixOS/nixpkgs
    ([8c13320](https://git.alanpearce.eu/searchix/commit/?id=8c1332020d25f74baa463bb1cec0e6783f565034))
-   render markdown examples
    ([c0fbf11](https://git.alanpearce.eu/searchix/commit/?id=c0fbf11f843af84e8891a708c4d217dd6c523473))
-   link to nixpkgs on github for declarations
    ([71de547](https://git.alanpearce.eu/searchix/commit/?id=71de5479bc625aa101f72a24b056b687b0630a9a))
-   auto-linkify URLs in markdown
    ([0d65324](https://git.alanpearce.eu/searchix/commit/?id=0d653242b21abd1678df35e2ba6c30cc23733eb1))
-   render markdown in option descriptions
    ([2430f46](https://git.alanpearce.eu/searchix/commit/?id=2430f46a9948b38b06880606a95dec357d01f619))
-   browser history management
    ([87ec9ec](https://git.alanpearce.eu/searchix/commit/?id=87ec9ecf21781c5289257750fd41c6b9991f1f6e))
-   show metadata in grid
    ([b682ed9](https://git.alanpearce.eu/searchix/commit/?id=b682ed936d8185e91331e9f66fc0e13072f8b545))
-   frontend search implementation
    ([5b9e67f](https://git.alanpearce.eu/searchix/commit/?id=5b9e67fd5129dec75169a1a070c70f910dff6da2))
-   create nixos options.json importer
    ([1fb64b0](https://git.alanpearce.eu/searchix/commit/?id=1fb64b09e2dba34784732a61cbebc119cd847803))
-   add basic input frontend
    ([0945da8](https://git.alanpearce.eu/searchix/commit/?id=0945da8f9654252008ff6bd56022f9363fd3cca6))
-   add css
    ([177c2de](https://git.alanpearce.eu/searchix/commit/?id=177c2de3f7c2b1cee8dbea4acaa896a334ba3162))
-   watch and live reload in development
    ([8411fb4](https://git.alanpearce.eu/searchix/commit/?id=8411fb4a3acebe46acaf7a2ff6c4e58018737d65))
-   serve a very basic html template
    ([7360307](https://git.alanpearce.eu/searchix/commit/?id=73603079e29bc89c54296a9e12b5a779cd84c023))

### Fixes

-   render errors as HTML given correct content-type
    ([2edf3f0](https://git.alanpearce.eu/searchix/commit/?id=2edf3f098dd1bb79fd97a8f30fe28938cbde9f3c))
-   missing last page of search results
    ([f9b0340](https://git.alanpearce.eu/searchix/commit/?id=f9b034027565bbc59b36d1a318716088c3a77977))
-   store only data from current entity in index
    ([3b877b3](https://git.alanpearce.eu/searchix/commit/?id=3b877b39b424ded0f776d1dc71d07616a47bfbd4))
-   nil pointer when updating all sources
    ([041ebd2](https://git.alanpearce.eu/searchix/commit/?id=041ebd2291db60e73d7fc213cd8a74d33b7f089b))
-   don't panic when re-indexing if fetcher did not update
    ([43fe431](https://git.alanpearce.eu/searchix/commit/?id=43fe431c3aff36d52a630ca670f9fc4ee161962f))
-   **css:** support browsers without css nesting
    ([29f2099](https://git.alanpearce.eu/searchix/commit/?id=29f2099e68a3226060308c66e5b1aa19fb2ba8e0))
-   **logging:** track whether headers are written properly
    ([c5c9f72](https://git.alanpearce.eu/searchix/commit/?id=c5c9f726cae4cf4488312338da9599da53ac8b6c))
-   rebuild index when --replace is specified
    ([d558039](https://git.alanpearce.eu/searchix/commit/?id=d558039919b6198a246a6a3fd007276191cb4b2f))
-   bad git conflict resolution
    ([d65d80e](https://git.alanpearce.eu/searchix/commit/?id=d65d80e1133eebf11cc1adb3438a1cc53635185f))
-   **flake:** remove overlay export
    ([8e22448](https://git.alanpearce.eu/searchix/commit/?id=8e22448e4dd53b284ba48bdf4ebf810bec489b41))
-   **importer:** don't create duplicate nested directories
    ([a94632e](https://git.alanpearce.eu/searchix/commit/?id=a94632eb788fd014bea0917a683896c44866f431))
-   **css:** barely-visible text in dark mode
    ([3fd832b](https://git.alanpearce.eu/searchix/commit/?id=3fd832b8f86641ea6046a776dda8f54a17834d7b))
-   **module:** include nix in path
    ([67e1a28](https://git.alanpearce.eu/searchix/commit/?id=67e1a28ad3972a31bb995eaeec83434221ed5cf6))
-   **config:** show read URL on parse failure
    ([45432ae](https://git.alanpearce.eu/searchix/commit/?id=45432ae8006322e89f89f43423ff26c99d98de10))
-   **module:** allow setting sentry DSN
    ([f700a32](https://git.alanpearce.eu/searchix/commit/?id=f700a3284ecc88bb2af1df5f0c8d242290a549cc))
-   don't panic on TOML decode error
    ([53c9fb2](https://git.alanpearce.eu/searchix/commit/?id=53c9fb2f72cbc82177404b76a4aad14283f8ced1))
-   linking to wrong declaration URLs
    ([df77930](https://git.alanpearce.eu/searchix/commit/?id=df77930bef039e4aef024574c60fb64d75287eee))
-   **import:** don't remove index if there are errors
    ([ef50bbf](https://git.alanpearce.eu/searchix/commit/?id=ef50bbf934a3024fdb82c31f5dc38e017603a86f))
-   **js:** make reload after error execute attempted search
    ([84acac6](https://git.alanpearce.eu/searchix/commit/?id=84acac6ddf46351d4c049c7bb25d72584b2c0fb7))
-   js error when no results found
    ([628c2a8](https://git.alanpearce.eu/searchix/commit/?id=628c2a858191cde34478b760b3c8c6c23a420088))
-   missing menu on search pages
    ([793c89a](https://git.alanpearce.eu/searchix/commit/?id=793c89a87cb58f09678bdf150694a369c088d914))
-   no results for home-manager search
    ([0eccb18](https://git.alanpearce.eu/searchix/commit/?id=0eccb18ab0a7241c1f6f5fee99a3961101e35392))
-   **config:** add default channel name for home-manager
    ([3554d8a](https://git.alanpearce.eu/searchix/commit/?id=3554d8afb600b2eb62fbe7c2d4792debb9ce7a60))
-   missing menu on search page
    ([42dd2f6](https://git.alanpearce.eu/searchix/commit/?id=42dd2f64f288a7474f356a8ae90e7409b7e38fb6))
-   pointer type mismatch
    ([8ab9adf](https://git.alanpearce.eu/searchix/commit/?id=8ab9adf719090e213bd91b2747986ba6ef94576d))
-   **js:** pushstate also on errors, so that the user can refresh
    ([453ae85](https://git.alanpearce.eu/searchix/commit/?id=453ae8569ab58fcc4ad61c461adc4489b9443cf8))
-   **js:** null value error for pagination
    ([83f6751](https://git.alanpearce.eu/searchix/commit/?id=83f6751139ed2dbf2f7c2f235e03d91d632e38bb))
-   incorrect default/example values stored after processing
    ([9f1a4aa](https://git.alanpearce.eu/searchix/commit/?id=9f1a4aaa9960fe3ab664033e1c1f42e62c3fb2e8))
-   incorrect data shown for default and example option keys
    ([5e008c2](https://git.alanpearce.eu/searchix/commit/?id=5e008c251cf0e91ca6cdf9d91faaaa4975636541))
-   panic on importing darwin options
    ([9b24b80](https://git.alanpearce.eu/searchix/commit/?id=9b24b805b45c71bb54dacc1112380357050f48af))
-   use options exposed via user's <nixpkgs> channel instead of ours
    ([1fdd1f8](https://git.alanpearce.eu/searchix/commit/?id=1fdd1f839782339f39a16ba3100dec129cd40a0c))
-   **css:** first grid row too tall
    ([3c54a3f](https://git.alanpearce.eu/searchix/commit/?id=3c54a3f63ad9b9ccaf5ca064ec3fd730164aad4a))

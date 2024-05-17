# Changelog

## Unreleased (2024-05-17)

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

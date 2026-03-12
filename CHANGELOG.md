# Changelog

## [v0.4.5](https://codeberg.org/alinnow/searchix/compare/v0.4.4...f2716d530792eaaf86e8b78de9baac9c945c35e0) (2026-03-11)

### Features

- Show search result number
  ([f2716d5](https://codeberg.org/alinnow/searchix/commit/f2716d530792eaaf86e8b78de9baac9c945c35e0))
- Add basic page meta description
  ([c3fa5a4](https://codeberg.org/alinnow/searchix/commit/c3fa5a442ee81a4c2252736c4d4ef461c6b3af93))
- make server graceful shutdown timeout configurable
  ([078020e](https://codeberg.org/alinnow/searchix/commit/078020ea68fd850a3d1dcfcdfbcb2ba61312161b))
- add page information to title
  ([9728e17](https://codeberg.org/alinnow/searchix/commit/9728e17d1b90757b74e9a15e41578b0db758b8c9))

### Fixes

- oddly-sized site title link hover background
  ([0227aae](https://codeberg.org/alinnow/searchix/commit/0227aaeb21d3640bd2b2a3b438a9bf38bbd54c20))
- Eval only in correct condition
  ([6cdd9ce](https://codeberg.org/alinnow/searchix/commit/6cdd9ced64d6f9d4189ecb2c30d028d3cc108054))
- **js:** close details dialog when using browser back action
  ([86f1002](https://codeberg.org/alinnow/searchix/commit/86f100286dbd86a2168bf6cda5d557042364b9dd))
- **js:** keep site title when adding page title
  ([911a857](https://codeberg.org/alinnow/searchix/commit/911a85738b3afef528d548a6b554c2f7d6ecb17e))
- panic on markdown output
  ([45a73d3](https://codeberg.org/alinnow/searchix/commit/45a73d31ca17a17cebf1d2c933c6d1b88c8d9ef4))

### [v0.4.4](https://codeberg.org/alinnow/searchix/compare/v0.4.3...v0.4.4) (2025-12-05)

#### Features

- add apple-touch-icon
  ([6001753](https://codeberg.org/alinnow/searchix/commit/6001753d2948ac6dddc4f2f7ba8911fcb45fde7d))

#### Fixes

- error opening details pages
  ([4cb44e8](https://codeberg.org/alinnow/searchix/commit/4cb44e8d904d3cd97570babd0dd05fb57355d5e7))
- error opening details pages
  ([4b263dc](https://codeberg.org/alinnow/searchix/commit/4b263dcbddee9a657ab10c5ae47e677c6bf67fcc))
- changing facets reverted to combined search
  ([560b635](https://codeberg.org/alinnow/searchix/commit/560b6350141050cd4b9aa5717acf5c241996a978))
- Add meta info for package to fix nix run usage
  ([29badc6](https://codeberg.org/alinnow/searchix/commit/29badc67e9722ac382d86382deabff0180280080))

### [v0.4.3](https://codeberg.org/alinnow/searchix/compare/v0.4.2...v0.4.3) (2025-11-12)

#### Features

- improve error logging of programs database
  ([bf67263](https://codeberg.org/alinnow/searchix/commit/bf67263fd74134bbe39d6de3f693b89937646bdc))
- remove package name from results table
  ([8ec4832](https://codeberg.org/alinnow/searchix/commit/8ec4832bfece859ba62a4092e074b535601aab18))

### [v0.4.2](https://codeberg.org/alinnow/searchix/compare/v0.4.1...v0.4.2) (2025-11-08)

#### Features

- improve results/facets display on narrow screens
  ([3a04aef](https://codeberg.org/alinnow/searchix/commit/3a04aefd4233216ef4c56dbc5befcc5b102f5201))
- add header to facets panel
  ([8acfcb9](https://codeberg.org/alinnow/searchix/commit/8acfcb9f668152007376c851be9caf93fe455096))

#### Fixes

- don't show last index time if it never finished
  ([48c28a8](https://codeberg.org/alinnow/searchix/commit/48c28a808535eb62d7bce4ddf3275fcd3b8e3191))

### [v0.4.1](https://codeberg.org/alinnow/searchix/compare/v0.4.0...v0.4.1) (2025-11-06)

#### Features

- don't block startup for indexing
  ([d8581c5](https://codeberg.org/alinnow/searchix/commit/d8581c5d9b41485c6be7cdb61cdfca4ba323e593))

#### Fixes

- negative waitgroup counter panic
  ([7971216](https://codeberg.org/alinnow/searchix/commit/7971216b5bd8d6df7f4a27379b16b783c5ce0132))

## [v0.4.0](https://codeberg.org/alinnow/searchix/compare/v0.3.0...v0.4.0) (2025-11-06)

### ⚠ BREAKING CHANGE

- requires `--outdir` instead of writing to stdout

### Fixes

- use correct search name for combined search
  ([0d664b1](https://codeberg.org/alinnow/searchix/commit/0d664b10b50db1871770826fd09171575a712595))
- make generate-error-page command output assets
  ([d29ccd9](https://codeberg.org/alinnow/searchix/commit/d29ccd9bebef309aea3dfddea1dfbd1f62ca38a5))

## [v0.3.0](https://codeberg.org/alinnow/searchix/compare/v0.2.8...v0.3.0) (2025-11-05)

### Features

- use plurals for document IDs and URLs
  ([ff28328](https://codeberg.org/alinnow/searchix/commit/ff283282d42ec5f77e9ad8f2c5616e4315b0355e))
- force-recreate index when schema version updated
  ([5f74928](https://codeberg.org/alinnow/searchix/commit/5f74928b831c3a82cf343d397929a08374d3beeb))

### Fixes

- record meta existence at startup
  ([35c0945](https://codeberg.org/alinnow/searchix/commit/35c0945e328d09292281439a280b5fce5d172a31))
- missing index step when new sources added to configuration
  ([dedd43f](https://codeberg.org/alinnow/searchix/commit/dedd43f7a653ccf3ccda11c9a6f34bbc227aa484))
- make error text readable in in dark mode
  ([bfa2851](https://codeberg.org/alinnow/searchix/commit/bfa2851e2a6d1800569ea5c451242bdf7f96dcac))
- enable wrapping of long option names in facets
  ([7975430](https://codeberg.org/alinnow/searchix/commit/7975430794939d3be24ed6f52856ac29555d5900))
- 404 on detail pages
  ([0f1f8e6](https://codeberg.org/alinnow/searchix/commit/0f1f8e6b60a0aa8d615d1b4838abd71faf045c64))

### [v0.2.8](https://codeberg.org/alinnow/searchix/compare/v0.2.7...v0.2.8) (2025-11-04)

#### Features

- re-enable immutable assets
  ([179fdc3](https://codeberg.org/alinnow/searchix/commit/179fdc3b417869d50a181cac813e1dbbf2893478))
- add 'parents' facet to options
  ([ab21d46](https://codeberg.org/alinnow/searchix/commit/ab21d46e298bcf5caf5ef50a12127df89f8caa5f))
- support faceted navigation of search results
  ([82cfb98](https://codeberg.org/alinnow/searchix/commit/82cfb98df7f9e38f295d3ea1d46b343914947307))

#### Fixes

- assets not being rehashed on change in development
  ([ee5cd9a](https://codeberg.org/alinnow/searchix/commit/ee5cd9a58f2e38efacec9b91f7208ba0f81d28a0))
- make treefmt's nixpkgs input follow project nixpkgs
  ([a716284](https://codeberg.org/alinnow/searchix/commit/a716284145943c7da204a9278c06772649b163a7))

### [v0.2.7](https://codeberg.org/alinnow/searchix/compare/v0.2.6...v0.2.7) (2025-10-17)

#### Features

- add debug logging for prune search
  ([3348e3f](https://codeberg.org/alinnow/searchix/commit/3348e3f8f5d151d8a81da8f2b9540c8dd0d1184e))
- revert back to using index as storage
  ([7c0e372](https://codeberg.org/alinnow/searchix/commit/7c0e3729dd314a571fd635408fb89f24199e00b3))
- add icon to OpenSearch Description XML
  ([5cc6ab7](https://codeberg.org/alinnow/searchix/commit/5cc6ab750051e5afc26b945327c73709de20b107))
- serve assets from root
  ([60b8ef0](https://codeberg.org/alinnow/searchix/commit/60b8ef0aad7364f462449990e9edd7b53d01c4be))
- add favicon
  ([43ad85a](https://codeberg.org/alinnow/searchix/commit/43ad85ad7c45ea39eeca1721112ff109c1614694))

#### Fixes

- save metadata after import
  ([fce18f9](https://codeberg.org/alinnow/searchix/commit/fce18f928e4f085875ee714a8bb0549388d6e693))
- incorrect javascript linking
  ([e0bbccf](https://codeberg.org/alinnow/searchix/commit/e0bbccf0b9c5e43bfa2ef02a5bb33c27b8bf5d00))

### [v0.2.6](https://codeberg.org/alinnow/searchix/compare/v0.2.5...v0.2.6) (2025-10-06)

#### Features

- restart instead of stopping service during generation switches
  ([d94a199](https://codeberg.org/alinnow/searchix/commit/d94a199425050ef9b1da09e5b8bba7611d3c1082))
- use sd_notify for watchdog and status notification
  ([49031e3](https://codeberg.org/alinnow/searchix/commit/49031e3ce0a01afda5079015c442a67720017220))

#### Fixes

- incorrectly selecting too many entities to prune
  ([eeac61c](https://codeberg.org/alinnow/searchix/commit/eeac61c274b3ee615aedf3318c97715d70915a72))
- avoid listener error on shutdown
  ([6d79f11](https://codeberg.org/alinnow/searchix/commit/6d79f1131c25a59f6ea7f4957d12813f8c62e73e))
- CSS not allowed by subresource integrity (disable for now)
  ([c4dc9b2](https://codeberg.org/alinnow/searchix/commit/c4dc9b20cce2b08e7cf0f0e0b6c21fb3938177d0))

### [v0.2.5](https://codeberg.org/alinnow/searchix/compare/v0.2.4...v0.2.5) (2025-09-03)

#### Features

- also prune storage and index on SIGUSR1
  ([afc540f](https://codeberg.org/alinnow/searchix/commit/afc540fd97b14f96cac85e6d45de669f6d35a00d))
- improve logging of pruning process
  ([dd9ca0b](https://codeberg.org/alinnow/searchix/commit/dd9ca0bf112a847749f4b92276baa4b89b153b8a))
- improve logging of errors in source ID handler
  ([7f75159](https://codeberg.org/alinnow/searchix/commit/7f7515906b28675a38a9988eb25f9b5b527c9d95))
- include search query in error messages
  ([77a8dec](https://codeberg.org/alinnow/searchix/commit/77a8dec1d8c80ff22bc539f7bd74bf20ad031353))
- make manual fetch always update sources
  ([ff56604](https://codeberg.org/alinnow/searchix/commit/ff56604e64988b29c6557f57ccf59284747843ec))

#### Fixes

- abort pruning if too many entities selected
  ([6ae7560](https://codeberg.org/alinnow/searchix/commit/6ae7560fb293ff0bc1dbd2cc490040ac2990266e))

### [v0.2.4](https://codeberg.org/alinnow/searchix/compare/v0.2.3...v0.2.4) (2025-08-25)

#### Features

- **storage:** respect --replace flag
  ([abef2e5](https://codeberg.org/alinnow/searchix/commit/abef2e5b0db37c771464469c3f83890b4cc0a297))
- improve error logging during import
  ([bc04964](https://codeberg.org/alinnow/searchix/commit/bc04964c3405f9d50bbad7193b61ed62d9e64498))

#### Fixes

- not re-indexing despite --reindex argument
  ([5afd91c](https://codeberg.org/alinnow/searchix/commit/5afd91c0b78982e8d1dfd41a911cf85d108979d9))

### [v0.2.3](https://codeberg.org/alinnow/searchix/compare/v0.2.2...v0.2.3) (2025-08-25)

#### Features

- re- ingest/index on SIGUSR1/2
  ([832214d](https://codeberg.org/alinnow/searchix/commit/832214d72d377f7c02146a77528b6e9c6bb393c8))
- listen for SIGHUP if running in a terminal
  ([15a5b32](https://codeberg.org/alinnow/searchix/commit/15a5b3210a073c83dbd1ace8969df83c53dac755))

#### Fixes

- always re-index if storage is (re-)created
  ([12a29a5](https://codeberg.org/alinnow/searchix/commit/12a29a55fe895a964d8d317857ee7fb87d4071f1))
- invalid config generated by nixos module
  ([0da8640](https://codeberg.org/alinnow/searchix/commit/0da8640705714335a1200e84287ff22e3fe7f5d6))
- unable to set log level from nixos module
  ([6d42ad1](https://codeberg.org/alinnow/searchix/commit/6d42ad10e483590d1dbb7ceacae8f6dd11f0d1df))
- add missing prefix to version link in footer
  ([5c14549](https://codeberg.org/alinnow/searchix/commit/5c1454971ad761d7ee56a4d9a8b33f1e789dc9e3))

### [v0.2.2](https://codeberg.org/alinnow/searchix/compare/v0.2.1...v0.2.2) (2025-08-21)

#### Features

- disable storm internal logging
  ([180d4cf](https://codeberg.org/alinnow/searchix/commit/180d4cf8780230e5af64e1750a2b2e4b87d519d3))
- make web search timeout configurable
  ([8a1972a](https://codeberg.org/alinnow/searchix/commit/8a1972abb4b429bd62b47cb44e3219d722919faf))

#### Fixes

- error setting up programs DB
  ([3be8a83](https://codeberg.org/alinnow/searchix/commit/3be8a83d4c44f7e21277abde2da69b54111c1c1a))
- enable graceful shutdown
  ([ff3d5a3](https://codeberg.org/alinnow/searchix/commit/ff3d5a36af52cf8bb06bf9c0e3a519536e2fa2f2))

### [v0.2.1](https://codeberg.org/alinnow/searchix/compare/v0.2.0...v0.2.1) (2025-08-13)

#### Features

- remove entities not present in current import
  ([d6d0bf3](https://codeberg.org/alinnow/searchix/commit/d6d0bf3856234f06ca5707df6c760904255bef08))
- pretty-print schema metadata file
  ([85276e9](https://codeberg.org/alinnow/searchix/commit/85276e96acd75dce54d0295dcb728cf86f096a45))
- store import timestamp on entities
  ([35226fd](https://codeberg.org/alinnow/searchix/commit/35226fd9989a28f813380e665a026caf427cb549))

#### Fixes

- storage error
  ([f896b68](https://codeberg.org/alinnow/searchix/commit/f896b6861faae07485b8f7bee2e9b607527882c1))
- error when exporting from storage to index
  ([131ccd3](https://codeberg.org/alinnow/searchix/commit/131ccd3edc02facbceb27c4ac44451a45bbb384e))

## [v0.2.0](https://codeberg.org/alinnow/searchix/compare/v0.1.46...v0.2.0) (2025-08-13)

### ⚠ BREAKING CHANGE

- searchix-web requires `serve` argument

### Features

- add check-config command
  ([6196d25](https://codeberg.org/alinnow/searchix/commit/6196d25ad331bdcbd0adc3da6facbcadc9bd3a1e))
- split storage import and indexing
  ([40d106a](https://codeberg.org/alinnow/searchix/commit/40d106a50be7fa87739e82c87c161db8e6a69b12))
- store data separate from search index
  ([520dcca](https://codeberg.org/alinnow/searchix/commit/520dcca2755dbbab13ceebdc53d03beec8b94533))
- add offline flag to reindex (mostly for dev)
  ([21e9e77](https://codeberg.org/alinnow/searchix/commit/21e9e778cc6be3c8717cd4286bedd17f2b22282a))
- improve tracking and display of import runs
  ([faf67ad](https://codeberg.org/alinnow/searchix/commit/faf67ad3fe2514492c6e0b0cf5361d5b44275a06))
- add --fetch action to populate data/sources
  ([001c5b4](https://codeberg.org/alinnow/searchix/commit/001c5b4f4e6dcf0110504211d3f1298d5e6abc48))
- make downloaded filenames consistent
  ([a49d405](https://codeberg.org/alinnow/searchix/commit/a49d405f81ed9b872ae638832f3a1a5f2339780c))
- store fetched files in data directory
  ([bdb5a54](https://codeberg.org/alinnow/searchix/commit/bdb5a54c661407c39668096074a1f4a57898eb77))
- limit file operations using os.Root
  ([1d518f4](https://codeberg.org/alinnow/searchix/commit/1d518f42e04712c84dfc168cc7a286aabb56e2ed))
- switch from errors to fault
  ([3dfbd8d](https://codeberg.org/alinnow/searchix/commit/3dfbd8dd7212f74622ba4892efb34bf3487da09b))

### Fixes

- update search index as part of update timer
  ([a1bbe1c](https://codeberg.org/alinnow/searchix/commit/a1bbe1c0565acf322ca9af88876b9d0c8ee52ab3))
- styles broken after updates due to stale cache/hash mismatch
  ([450ff2d](https://codeberg.org/alinnow/searchix/commit/450ff2d51460c2f1e8b0679726d3a26cd8762359))
- handle missing manpage URL map without aborting
  ([a34ba86](https://codeberg.org/alinnow/searchix/commit/a34ba86c2c899bbf2902c4c99b817a9692c3294f))
- set date of files copied from nix store
  ([762d97b](https://codeberg.org/alinnow/searchix/commit/762d97b5cca3a5b1605f34bff57e785b93833450))
- don't try to parse an empty last-modified header
  ([cee1276](https://codeberg.org/alinnow/searchix/commit/cee12761d4ab83d717452e22a34708161f722b1f))
- error saving metadata when re-indexing
  ([3e92836](https://codeberg.org/alinnow/searchix/commit/3e928369c3af69d0ef24a2f5d20c938689b15aa1))
- don't log and exit on graceful shutdown timeout
  ([1db432a](https://codeberg.org/alinnow/searchix/commit/1db432a56d503d60ccfafadfe6494f64cb8c325c))
- import package even with incorrect 'homePage' casing
  ([4c66828](https://codeberg.org/alinnow/searchix/commit/4c6682878224589ec6c8b14cbe03eefb61e4d328))
- re-enable margin-top for headers in full page view
  ([7b08d54](https://codeberg.org/alinnow/searchix/commit/7b08d540424f61cdf1d379f7d60e15641822c86e))

### [v0.1.46](https://codeberg.org/alinnow/searchix/compare/v0.1.45...v0.1.46) (2025-07-07)

#### Features

- enable advanced query string syntax
  ([465a9a8](https://codeberg.org/alinnow/searchix/commit/465a9a8d03878f9b8f22424d679080fb53317d6d))
- improve matching of multiple keywords
  ([516f3e9](https://codeberg.org/alinnow/searchix/commit/516f3e9dcd9b987e6ddda7741f088901b319fe38))

### [v0.1.45](https://codeberg.org/alinnow/searchix/compare/v0.1.44...v0.1.45) (2025-07-07)

#### Features

- add link to status page
  ([35aa773](https://codeberg.org/alinnow/searchix/commit/35aa773dd51342bc23fbafcc3dcceab9bcd9387e))

#### Fixes

- set default values for custom sources
  ([756d8f1](https://codeberg.org/alinnow/searchix/commit/756d8f1f343473e7e33b81eceb11e2bb608752b6))
- avoid panic when rendering combined data
  ([1685f38](https://codeberg.org/alinnow/searchix/commit/1685f383414d5a184376519fa81002a506712534))
- panic rendering corrupted data
  ([a2b341d](https://codeberg.org/alinnow/searchix/commit/a2b341d554d9fecc5a4a90eec5438eb1dec2de04))

### [v0.1.44](https://codeberg.org/alinnow/searchix/compare/v0.1.43...v0.1.44) (2025-06-25)

#### Fixes

- descriptions truncated by one character
  ([e7d8aad](https://codeberg.org/alinnow/searchix/commit/e7d8aad2ed535ab5e887cd6aac9459696be9b3fe))
- remove trailing full-stop from single-sentence descriptions
  ([c102058](https://codeberg.org/alinnow/searchix/commit/c102058ee750d770115a47cbeda3116208009fbd))

### [v0.1.43](https://codeberg.org/alinnow/searchix/compare/v0.1.42...v0.1.43) (2025-06-12)

#### Fixes

- first sentence of some descriptions cut off
  ([f608afd](https://codeberg.org/alinnow/searchix/commit/f608afdaebe0fdbf49b8363ef5adf6f9f0e918cb))

### [v0.1.42](https://codeberg.org/alinnow/searchix/compare/v0.1.41...v0.1.42) (2025-06-12)

#### Features

- link man pages like how NixOS does it
  ([2f2d869](https://codeberg.org/alinnow/searchix/commit/2f2d86922ba23c4b7f92b115f4e8d26e5058bd23))
- parse {man} in markdown documentation
  ([813e5a2](https://codeberg.org/alinnow/searchix/commit/813e5a26b629d4c27b6ce0a8de2e3d04308ef535))
- parse {file} in markdown documentation
  ([a4c441f](https://codeberg.org/alinnow/searchix/commit/a4c441fe30ae008460149a06a5937ca37b4c2d72))
- parse {var} in markdown documentation
  ([d296272](https://codeberg.org/alinnow/searchix/commit/d2962720cfd4b4f022ccd564f39189db1ccfa0b7))
- parse {env} in markdown documentation
  ([7ab69fb](https://codeberg.org/alinnow/searchix/commit/7ab69fb0411fc260e3f3ea519443abd87ace3a7e))
- parse {command} in markdown documentation
  ([cb907ad](https://codeberg.org/alinnow/searchix/commit/cb907ad36521036db199a226f3e357ad3b7b0a1c))
- parse {option} in markdown documentation
  ([0ba69e5](https://codeberg.org/alinnow/searchix/commit/0ba69e50d2d1095db9b42c548c0226739d137a62))
- enable parsing of markdown fences
  ([2d958a5](https://codeberg.org/alinnow/searchix/commit/2d958a53b942e5c47ce58dd14c15a2c9474ca661))
- create option link parser
  ([93b800b](https://codeberg.org/alinnow/searchix/commit/93b800b4862d5c7e0e621647d78b79416a9f3101))

### [v0.1.41](https://codeberg.org/alinnow/searchix/compare/v0.1.40...v0.1.41) (2025-06-09)

#### Features

- improve logging of import process
  ([e0e35b1](https://codeberg.org/alinnow/searchix/commit/e0e35b1818a49ba006c500e292efe0b356b34a44))
- enable automatic re-indexing on schema version changes
  ([1d4eff7](https://codeberg.org/alinnow/searchix/commit/1d4eff79b408b95f2dde5db8aa3cec0a7f43adb4))

#### Fixes

- matching long prefixes in option names (regression)
  ([d2bc9be](https://codeberg.org/alinnow/searchix/commit/d2bc9be6ec207fb06418e56774f64034eee8b57f))
- don't quit after generating index for first time
  ([7514e61](https://codeberg.org/alinnow/searchix/commit/7514e6182832fc7806d2d45d32ea87ad3cdc668c))
- keep names_with_underscores as single tokens
  ([87f2c60](https://codeberg.org/alinnow/searchix/commit/87f2c60692a7ab6a46b969ddd29e1c97e0cd6722))

### [v0.1.40](https://codeberg.org/alinnow/searchix/compare/v0.1.39...v0.1.40) (2025-05-29)

#### Fixes

- exact/prefix matches for package name/attr not shown
  ([1f7e995](https://codeberg.org/alinnow/searchix/commit/1f7e995a669eefd259ab773492c3e5a673e2dac6))

### [v0.1.39](https://codeberg.org/alinnow/searchix/compare/v0.1.38...v0.1.39) (2025-05-29)

#### Fixes

- match option name prefixes even when len > 25
  ([61f28d3](https://codeberg.org/alinnow/searchix/commit/61f28d3acb0f3cc7cc37ee384453c9cdfbaa9dbb))

### [v0.1.38](https://codeberg.org/alinnow/searchix/compare/v0.1.37...v0.1.38) (2025-04-13)

#### Features

- improve display on smaller screens
  ([8efd1e2](https://codeberg.org/alinnow/searchix/commit/8efd1e2eccf9e894926a73b2800dffaf00886f7c))
- hide description table on smaller screens
  ([1b8b6b5](https://codeberg.org/alinnow/searchix/commit/1b8b6b506b0855afc282bc1cc7b631f02e97b8e3))

#### Fixes

- scroll header nav instead of wrapping words on tablets
  ([8d11809](https://codeberg.org/alinnow/searchix/commit/8d118090b0c195ccfe5cec9025274cbf8314e54d))

### [v0.1.37](https://codeberg.org/alinnow/searchix/compare/v0.1.36...v0.1.37) (2025-04-06)

#### Fixes

- package definition link regressed since v0.1.31
  ([0cceea8](https://codeberg.org/alinnow/searchix/commit/0cceea81794a9c6b0faff85d51d9602f4ac68513))
- missing base.css in package lib folder
  ([2b50775](https://codeberg.org/alinnow/searchix/commit/2b5077550069013bc51e86bbb41bf6f7dc2cbf71))

### [v0.1.36](https://codeberg.org/alinnow/searchix/compare/v0.1.35...v0.1.36) (2025-04-03)

#### Features

- prioritise whole-phrase matches
  ([c6bc629](https://codeberg.org/alinnow/searchix/commit/c6bc629f27ab2005c10005587eea1e28c946db55))
- add flag to generate error page for web server
  ([5127a98](https://codeberg.org/alinnow/searchix/commit/5127a98c591d206c0791b3049e8ae30b87ddf2dd))

#### Fixes

- prioritise exact attribute matches above all
  ([c2bbda8](https://codeberg.org/alinnow/searchix/commit/c2bbda87df1c708dc8437ef49b296c11ba251355))
- erroneously running importer on every startup
  ([e699eaa](https://codeberg.org/alinnow/searchix/commit/e699eaa3f57b7c67825aa9d1c507e03254f422be))

### [v0.1.35](https://codeberg.org/alinnow/searchix/compare/v0.1.34...v0.1.35) (2025-04-02)

#### Features

- enable searching homepages field
  ([0ceaf8e](https://codeberg.org/alinnow/searchix/commit/0ceaf8ed86c26ea4cb1c11ce6dfbfab88297a464))
- improve logging of next import time
  ([00b189d](https://codeberg.org/alinnow/searchix/commit/00b189dbba42f084ffb94e28b3b709bb3f65d32e))
- make index batch size configurable
  ([f38ccb5](https://codeberg.org/alinnow/searchix/commit/f38ccb5ec8149072c93a6c3173da06ba1d724c4c))
- list package platforms in detail view
  ([f8f4f74](https://codeberg.org/alinnow/searchix/commit/f8f4f748c8528b7fe245d1e1f1d55f78b78ede39))

#### Fixes

- importer timeout ignored in scheduler
  ([75c2afd](https://codeberg.org/alinnow/searchix/commit/75c2afdf214b0ddc57efe8e621eb172506c0b0f7))

### [v0.1.34](https://codeberg.org/alinnow/searchix/compare/v0.1.33...v0.1.34) (2025-03-24)

#### Features

- buffer fetched HTTP bodies
  ([abf819d](https://codeberg.org/alinnow/searchix/commit/abf819db104560670cdc388a584d09968aac7bbb))
- re-index on startup if last run > 24 hours ago
  ([6984d4d](https://codeberg.org/alinnow/searchix/commit/6984d4d32ab506494394f8a6a8aa18041c45b9e8))

#### Fixes

- another typo in search markup
  ([efaaf3f](https://codeberg.org/alinnow/searchix/commit/efaaf3f5296807eec377b19d4733fd360c8c655f))
- typo in indexing status
  ([fecc84f](https://codeberg.org/alinnow/searchix/commit/fecc84fab93bb3228be5fee4160820751195feec))
- nil pointer panic in importer
  ([0fb48cf](https://codeberg.org/alinnow/searchix/commit/0fb48cf6dd934778584120ebbc9e568c8b80b80e))

### [v0.1.33](https://codeberg.org/alinnow/searchix/compare/v0.1.32...v0.1.33) (2025-03-24)

#### Fixes

- build failure due to outdated file reference
  ([dd58d99](https://codeberg.org/alinnow/searchix/commit/dd58d9925b68c7cd639c5e217782709c1936497f))

### [v0.1.32](https://codeberg.org/alinnow/searchix/compare/v0.1.31...v0.1.32) (2025-03-24)

#### Features

- improve handling of long option descriptions
  ([4fa3ceb](https://codeberg.org/alinnow/searchix/commit/4fa3ceb5ac040a84f2da405c46e4af1231d22f17))
- widen body
  ([1c56ff7](https://codeberg.org/alinnow/searchix/commit/1c56ff7fc392f2d7992e53add5bd76bc8855d4a6))
- improve display of indexing times and durations
  ([297a7a2](https://codeberg.org/alinnow/searchix/commit/297a7a2096a12e6ac50d7b2077f2d75a7026fae4))
- link directly to version changelog in footer
  ([7d4ff28](https://codeberg.org/alinnow/searchix/commit/7d4ff28794fcc8d7dd8c87f300ce6a6235f87cc5))
- **js:** allow opening details in new tab/window
  ([03fa0eb](https://codeberg.org/alinnow/searchix/commit/03fa0eb00aef448bec05280a121c659a600fc13f))
- promote prefix/exact name/attr matches
  ([1328df9](https://codeberg.org/alinnow/searchix/commit/1328df9c5ca44e903a052c9bfc9a67dd34e25704))

#### Fixes

- missing content of last index run time
  ([0dc2635](https://codeberg.org/alinnow/searchix/commit/0dc2635e199591b917c8cc8ff40df3c3da6f9ec6))
- block single-character queries
  ([0cd9a24](https://codeberg.org/alinnow/searchix/commit/0cd9a2446413e9050c24ccd409bc8f97ad18076a))
- don't attempt to load more than configured number of results
  ([d39204e](https://codeberg.org/alinnow/searchix/commit/d39204eadf673b6c7ae940203fd75c0805245a96))

### [v0.1.31](https://codeberg.org/alinnow/searchix/compare/v0.1.30...v0.1.31) (2025-03-22)

#### Features

- **nixos-module:** allow setting environment variables
  ([e8fbdf3](https://codeberg.org/alinnow/searchix/commit/e8fbdf3bd12c8920a6e9bd84b34e787764b11eaf))
- **nixos-module:** allow setting environment variables
  ([78fc3e6](https://codeberg.org/alinnow/searchix/commit/78fc3e6cd43e6df90e17067fe0eb52b9badf9a75))
- demote NUR results in combined search
  ([782b636](https://codeberg.org/alinnow/searchix/commit/782b636c6ba9ebccecf46c7a1e5583e8108baf9a))
- promote results with literal or prefix name matches
  ([49a07fb](https://codeberg.org/alinnow/searchix/commit/49a07fb0b513dcaeb6241f4d87c160b9e0119260))
- make list of source links dynamic
  ([7247322](https://codeberg.org/alinnow/searchix/commit/7247322a386f065c643dc58f0ae5b57ad7ec1cc1))
- enable NUR package import
  ([2705e97](https://codeberg.org/alinnow/searchix/commit/2705e97ce1cf7d6a399c5f0175c36562fdef3352))
- wrap search form in semantic <search> element
  ([ff1e953](https://codeberg.org/alinnow/searchix/commit/ff1e9539fca1f011cfd52d0309a373f211c3fd10))
- show last/next/current indexing run time
  ([383ee78](https://codeberg.org/alinnow/searchix/commit/383ee780613116e78db9114a39a2d6127533463c))
- shorten shutdown timeout in development
  ([49e3004](https://codeberg.org/alinnow/searchix/commit/49e3004d33bf84aa081460e4a6d89a8d84cc12b0))
- Convert templ components to gomponents
  ([896d844](https://codeberg.org/alinnow/searchix/commit/896d844cac976afd0ee8aa73dd2fb28e15e7ac79))

#### Fixes

- remove rendering from search timeout restriction
  ([d4ec6e5](https://codeberg.org/alinnow/searchix/commit/d4ec6e5beecd549114dafd0b7c3b4a9d910388fb))
- package programs displayed off-centre
  ([de98780](https://codeberg.org/alinnow/searchix/commit/de987806cd030e85a22e11b35835a3524068adb7))
- footer spacing
  ([c0c02ac](https://codeberg.org/alinnow/searchix/commit/c0c02ac768a144f4417edfba967a4f7857a150b9))
- detach version from rest of footer text
  ([6b342b8](https://codeberg.org/alinnow/searchix/commit/6b342b83cedec82d240fc820d9696d3bb3eda8a2))
- wrong pagination links for combined results
  ([9102aef](https://codeberg.org/alinnow/searchix/commit/9102aef53c5fb73585359306a518e726a3623731))

### [v0.1.30](https://codeberg.org/alinnow/searchix/compare/v0.1.29...v0.1.30) (2025-03-18)

#### Features

- show version number and link in footer
  ([5762645](https://codeberg.org/alinnow/searchix/commit/5762645aedc4d39a9e6caeb227410ca9bae2d2b5))
- split compound words in names into n-grams
  ([b650b99](https://codeberg.org/alinnow/searchix/commit/b650b993e38f9919d2e65b732a5dee0267c71b84))

### [v0.1.29](https://codeberg.org/alinnow/searchix/compare/v0.1.28...v0.1.29) (2025-03-17)

#### Fixes

- nil pointer deference in nixpkgs-channel fetcher
  ([573e24b](https://codeberg.org/alinnow/searchix/commit/573e24bf30e5e8b8f5b82e848258b70eea9f3a74))
- build error (incomplete commit)
  ([d5e1147](https://codeberg.org/alinnow/searchix/commit/d5e11478bf3771623a0d181e6cede3145b3563e0))
- build error caused by incomplete commit
  ([f775db9](https://codeberg.org/alinnow/searchix/commit/f775db95d59a56feeafa6a4e147885c814bf5fb7))

### [v0.1.28](https://codeberg.org/alinnow/searchix/compare/v0.1.27...v0.1.28) (2025-03-12)

#### Fixes

- js error due to missing class on nav menu item
  ([9af2747](https://codeberg.org/alinnow/searchix/commit/9af2747a1168f41b39b446bfe3ecab6c7b25ff99))

### [v0.1.27](https://codeberg.org/alinnow/searchix/compare/v0.1.26...v0.1.27) (2025-03-11)

#### Fixes

- sources disabled by default in NixOS module
  ([333bd2f](https://codeberg.org/alinnow/searchix/commit/333bd2fbd16fc12c36e1b1d5b5bffdf065585fa8))

### [v0.1.26](https://codeberg.org/alinnow/searchix/compare/v0.1.25...v0.1.26) (2025-03-11)

#### Fixes

- nixos import fails with default settings
  ([5cdcac9](https://codeberg.org/alinnow/searchix/commit/5cdcac9f7a09b7cfd2fa699550092065b4d9a6c4))

### [v0.1.25](https://codeberg.org/alinnow/searchix/compare/v0.1.24...v0.1.25) (2025-03-11)

#### Features

- link to referenced projects
  ([cb37af4](https://codeberg.org/alinnow/searchix/commit/cb37af4c87525bffc9b6f045b85e78aab18bf49d))

#### Fixes

- panic when source order > number of sources
  ([d99fb7b](https://codeberg.org/alinnow/searchix/commit/d99fb7b6585837a02f43d1c3f37ed172baa1bf5d))

### [v0.1.24](https://codeberg.org/alinnow/searchix/compare/v0.1.23...v0.1.24) (2025-03-11)

#### Fixes

- **js:** switching source after reset uses old query
  ([12947e5](https://codeberg.org/alinnow/searchix/commit/12947e5a912ddcd6bb82cbaab2811d02de703871))

### [v0.1.23](https://codeberg.org/alinnow/searchix/compare/v0.1.22...v0.1.23) (2025-03-10)

#### Fixes

- remove duplicate link to source code
  ([1fe30ce](https://codeberg.org/alinnow/searchix/commit/1fe30ce514535c37ca668a0e426aa48362f0bdbf))

### [v0.1.22](https://codeberg.org/alinnow/searchix/compare/v0.1.21...v0.1.22) (2025-03-10)

#### Features

- default to combined search on root path
  ([05a32bd](https://codeberg.org/alinnow/searchix/commit/05a32bda78911772cb8b09f37bacbcf851a4ddba))

### [v0.1.21](https://codeberg.org/alinnow/searchix/compare/v0.1.20...v0.1.21) (2025-03-06)

#### Features

- prioritise exact matches on programs/attr/name field
  ([6c9a3a9](https://codeberg.org/alinnow/searchix/commit/6c9a3a93f0c55f65539230c4c979816df8e77c1b))
- improve relevance with compound words
  ([24e3fc9](https://codeberg.org/alinnow/searchix/commit/24e3fc96ce2c96401155b86c6bfb6a6fb2e322f8))

#### Fixes

- re-index refusing to delete data directory with programs.db
  ([3ee8248](https://codeberg.org/alinnow/searchix/commit/3ee8248656c089594ea8a3be41945c408ab39aa4))

### [v0.1.20](https://codeberg.org/alinnow/searchix/compare/v0.1.19...v0.1.20) (2025-01-24)

#### Fixes

- build failure (bad rebase)
  ([2d6315f](https://codeberg.org/alinnow/searchix/commit/2d6315fa9386aaecbc1a7fc6a7ffcceb0ae09c35))

### [v0.1.19](https://codeberg.org/alinnow/searchix/compare/v0.1.18...v0.1.19) (2025-01-24)

#### Fixes

- import slowdown when querying programs database
  ([1e50453](https://codeberg.org/alinnow/searchix/commit/1e5045349f93298b4c021c96c7c9c5a188489f2e))

### [v0.1.18](https://codeberg.org/alinnow/searchix/compare/v0.1.17...v0.1.18) (2025-01-20)

#### Features

- allow HTTP request logging to be disabled
  ([584d295](https://codeberg.org/alinnow/searchix/commit/584d295c25697c19ffd369d22281f7d53a139f18))
- **sentry:** trace all transactions
  ([19d8046](https://codeberg.org/alinnow/searchix/commit/19d8046f99d9f395f8529886f37292d99dc1a4d9))

### [v0.1.17](https://codeberg.org/alinnow/searchix/compare/v0.1.16...v0.1.17) (2025-01-19)

#### Fixes

- **sentry:** report correct HTTP path in traces
  ([eece88d](https://codeberg.org/alinnow/searchix/commit/eece88d32b1b613aba7ec2f11bcfeb10163616b1))
- use importer configured timeout for sentry check-ins
  ([b706b5a](https://codeberg.org/alinnow/searchix/commit/b706b5a3d09983637ea6d5af515733d77b674bdb))

### [v0.1.16](https://codeberg.org/alinnow/searchix/compare/v0.1.15...v0.1.16) (2025-01-15)

#### Features

- enable searching via program names for multi-program packages
  ([7aea6aa](https://codeberg.org/alinnow/searchix/commit/7aea6aa210a8939ac208fb7540d1b46ba69a995f))

#### Fixes

- **dev:** fix CSP not working in dev mode behind reverse proxy
  ([b26ddba](https://codeberg.org/alinnow/searchix/commit/b26ddba432f8bde78022d2fc8837f0ffb25448b1))

### [v0.1.15](https://codeberg.org/alinnow/searchix/compare/v0.1.14...v0.1.15) (2025-01-05)

#### Features

- offer to show all results on one page (if < 10000)
  ([9207642](https://codeberg.org/alinnow/searchix/commit/920764211064521930441b96ed914034302a7470))

#### Fixes

- **dev:** overwrite previous asset hash on change
  ([d289e54](https://codeberg.org/alinnow/searchix/commit/d289e54af28de6b21cbe49cebc262c4e3ca9f6d2))

### [v0.1.14](https://codeberg.org/alinnow/searchix/compare/v0.1.13...v0.1.14) (2025-01-02)

#### Features

- show scores in results table in dev mode
  ([58643f3](https://codeberg.org/alinnow/searchix/commit/58643f38bdd62dc538a03865b1f81eb77d957779))

#### Fixes

- boost non-stemmed matches over stemmed matches
  ([797c6e2](https://codeberg.org/alinnow/searchix/commit/797c6e27e18f6c47ceb4266d747a05064365c32d))

### [v0.1.13](https://codeberg.org/alinnow/searchix/compare/v0.1.12...v0.1.13) (2024-12-30)

#### Features

- increase results per page
  ([e1787bf](https://codeberg.org/alinnow/searchix/commit/e1787bf0b6dbac3a1601559d2f00ebc4b2df0f29))
- de-rank packages on 'all' search
  ([df5db96](https://codeberg.org/alinnow/searchix/commit/df5db969262db1382dcd9310d362efdf8f8b6949))

### [v0.1.12](https://codeberg.org/alinnow/searchix/compare/v0.1.11...v0.1.12) (2024-12-04)

#### Fixes

- use UTC time for fetch/import timer
  ([c92530b](https://codeberg.org/alinnow/searchix/commit/c92530b7940bfd9e0940dd07e4a33b8dc4b575ea))
- enable using module in pure evaluation mode
  ([ec4946e](https://codeberg.org/alinnow/searchix/commit/ec4946ee959b2d7d28287e9cd4643a0698833f6b))
- enable using module in pure evaluation mode
  ([174ad04](https://codeberg.org/alinnow/searchix/commit/174ad04905da9709e4aef366e14aa438e58e5b0b))

### [v0.1.11](https://codeberg.org/alinnow/searchix/compare/v0.1.10...v0.1.11) (2024-11-20)

#### Features

- apply stemming to search index terms
  ([72d56a6](https://codeberg.org/alinnow/searchix/commit/72d56a6781e97cde13ba068e8553e9342bd1fa20))

### [v0.1.10](https://codeberg.org/alinnow/searchix/compare/v0.1.9...v0.1.10) (2024-11-10)

#### Fixes

- search all missing expected results
  ([8656f47](https://codeberg.org/alinnow/searchix/commit/8656f47cd62b084a59546c79d437a120869661e3))

### [v0.1.9](https://codeberg.org/alinnow/searchix/compare/v0.1.8...v0.1.9) (2024-08-17)

#### Fixes

- don't error when item not found
  ([b427c26](https://codeberg.org/alinnow/searchix/commit/b427c2657582caa9b2cce8bcfd82624663dda780))
- tell browsers not to capitalise search input
  ([faf4d0e](https://codeberg.org/alinnow/searchix/commit/faf4d0e5f21266bdad99c991efef4e83c09da5c6))

### [v0.1.8](https://codeberg.org/alinnow/searchix/compare/v0.1.7...v0.1.8) (2024-07-13)

#### Features

- enable combined search
  ([6781684](https://codeberg.org/alinnow/searchix/commit/6781684647ad7f131054a93e1d13105d865c0e83))
- more structured logging
  ([94b21b2](https://codeberg.org/alinnow/searchix/commit/94b21b286edff37496a2fe481963625ac01c30a1))

#### Fixes

- nil panics rendering last page of search results
  ([b1fd710](https://codeberg.org/alinnow/searchix/commit/b1fd7104004b9646f541afbc0add3047dbff59a7))
- panic rendering error pages
  ([ee83dfd](https://codeberg.org/alinnow/searchix/commit/ee83dfd741c47bd25eac1ac7139f20875de0452c))
- opensearch searches not using input query
  ([a5f58bc](https://codeberg.org/alinnow/searchix/commit/a5f58bc7fe205b8ff84f6b654513e40bfbc61ae9))
- extra data after error parsing query string
  ([4423458](https://codeberg.org/alinnow/searchix/commit/442345898802e93b0b884fd0f1c3492e08c34049))
- panic rendering search results
  ([5a2d54d](https://codeberg.org/alinnow/searchix/commit/5a2d54d9c8751935284293b2ffd396fff0a5d202))
- panic on search error
  ([f076b5b](https://codeberg.org/alinnow/searchix/commit/f076b5bd6cb82edd99be50f3dbdd39bb9be2c44e))

### [v0.1.7](https://codeberg.org/alinnow/searchix/compare/v0.1.6...v0.1.7) (2024-06-29)

#### Features

- add message for users with javascript disabled
  ([e500a10](https://codeberg.org/alinnow/searchix/commit/e500a1089797eee1b027d5237b2e4f140004098c))

#### Fixes

- blank modal when using browser navigation
  ([8ef4bc9](https://codeberg.org/alinnow/searchix/commit/8ef4bc9b1f16dd6395d0e0daa85ada9b8db12015))
- duplicated results tables via JS search
  ([e0d86be](https://codeberg.org/alinnow/searchix/commit/e0d86bec7c9e8bedd1e71173b486668e16ec7c2e))
- panic rendering results
  ([dbeddce](https://codeberg.org/alinnow/searchix/commit/dbeddced4046cb5756af2604f5cb6734b3e7cee6))
- incorrect formatting of options results headers
  ([107d54a](https://codeberg.org/alinnow/searchix/commit/107d54a4c4a415f3581a452878cdd625f7f2634a))

### [v0.1.6](https://codeberg.org/alinnow/searchix/compare/v0.1.5...v0.1.6) (2024-06-19)

#### Features

- add low-memory mode
  ([63599c7](https://codeberg.org/alinnow/searchix/commit/63599c719d8e4e09548546bb366a849ef40c602b))
- enable configuration of flags as environment variables
  ([a495a67](https://codeberg.org/alinnow/searchix/commit/a495a673d2dda7b2aa9b11217d4662367a627517))
- display results in a table, showing details on click
  ([408aed0](https://codeberg.org/alinnow/searchix/commit/408aed03d3454330120475ca53838a2f4fe28ea3))
- add detail pages for packages/options
  ([d40c0e1](https://codeberg.org/alinnow/searchix/commit/d40c0e188a7fe1b36887f59c4a9958faa81b3d44))
- use stale-while-revalidate instead of immutable asset paths
  ([872ae5d](https://codeberg.org/alinnow/searchix/commit/872ae5db32c605ef385588a33673ce5fef29ff21))
- serve assets via immutable paths
  ([896926a](https://codeberg.org/alinnow/searchix/commit/896926a63a8f2d145467b325f9b0198315e0af6d))
- make security headers stricter
  ([b537694](https://codeberg.org/alinnow/searchix/commit/b53769462bf830f860b7d741a3d0801afdbc9aa2))
- enable sub-resource integrity for assets
  ([4698a97](https://codeberg.org/alinnow/searchix/commit/4698a97974ae82e7bd8592828c58294b222a58ff))

#### Fixes

- create index metadata when index exists and meta doesn't
  ([e665146](https://codeberg.org/alinnow/searchix/commit/e6651469b9e934cb4dc65f760277111a79e9ae2d))
- moving through search history shows no results
  ([66b2556](https://codeberg.org/alinnow/searchix/commit/66b2556a6a7c911a69b231fddeefe0a939d8898d))
- nil pointer exception when (re-)indexing
  ([2c4a19a](https://codeberg.org/alinnow/searchix/commit/2c4a19ad6cb71065d31304370708c859b4f68f6d))

### [v0.1.5](https://codeberg.org/alinnow/searchix/compare/v0.1.4...v0.1.5) (2024-05-29)

#### Features

- improve search relevance by boosting keyword matches
  ([73185e4](https://codeberg.org/alinnow/searchix/commit/73185e4633d4f56ce8acdd3cc19f78faaa78bd2b))
- support cleartext HTTP/2 (i.e. from a reverse proxy)
  ([16b4337](https://codeberg.org/alinnow/searchix/commit/16b433791b619f4144110e4429d98684bf5e98eb))

#### Fixes

- missing alternate forms of packages with the same name
  ([838e178](https://codeberg.org/alinnow/searchix/commit/838e178e503d9576a40fce8fc986c4e1972e4c0f))
- incorrect (switched) behaviour of --update/--replace
  ([97021f7](https://codeberg.org/alinnow/searchix/commit/97021f769a0f06103b4af5019d942759d644e291))
- re-enable exit --update/--replace
  ([a5e3c14](https://codeberg.org/alinnow/searchix/commit/a5e3c144034eab0cb41c4b65557e5c30e877c9fb))
- check for nil pointers in NewMux
  ([2c1491d](https://codeberg.org/alinnow/searchix/commit/2c1491de56d0c3e2f4cb0b0c1e33035510f72fc5))

### [v0.1.4](https://codeberg.org/alinnow/searchix/compare/v0.1.3...v0.1.4) (2024-05-23)

#### Features

- stream files directly from fetcher to importer
  ([0dbfe37](https://codeberg.org/alinnow/searchix/commit/0dbfe37fbddb95c184d845c79bbe014597d55fe8))
- add --update flag to force-update index and exit
  ([9c790c1](https://codeberg.org/alinnow/searchix/commit/9c790c1c620e46ca8d2cf49c50951635621e9a7d))
- **config:** use enums' string representations to print defaults
  ([a90419a](https://codeberg.org/alinnow/searchix/commit/a90419aa46ec42588f65e6327559ce479f9c5b67))
- **config:** print durations and URLs with human values
  ([0f55080](https://codeberg.org/alinnow/searchix/commit/0f5508013d776f8806a87957f80e74c7f74c72eb))
- show default configuration with --print-default-config
  ([60a1569](https://codeberg.org/alinnow/searchix/commit/60a15699657cef86d1a644c0d13b0d117a818633))

#### Fixes

- abort import of source on batch processing errors
  ([3053e41](https://codeberg.org/alinnow/searchix/commit/3053e41b1528ef898cccd44e056e4d167619af6b))
- **web:** fetch host from request, not header, where it's expected
  ([e69d785](https://codeberg.org/alinnow/searchix/commit/e69d78542a4ec92da5931b3db105b14a39a4c964))
- **nix:** generate default baseURL from listen address and port
  ([027db43](https://codeberg.org/alinnow/searchix/commit/027db4372e82e445c531c958c2da0f9cf6ae1e03))
- make fetcher check on creation that it supports Source.Importer
  ([b77a24f](https://codeberg.org/alinnow/searchix/commit/b77a24f9f75378ffe97be83cf4dfd7f1683b9a7e))
- **importer:** abort import of current source in case of error
  ([c821da1](https://codeberg.org/alinnow/searchix/commit/c821da1cf55864852bcd8f337dd7acd0cc02f0b9))
- nix cannot read date/time TOML values from default config
  ([202317d](https://codeberg.org/alinnow/searchix/commit/202317df1dba891654bf749c3720fbd602df04ff))
- duplicate error logging on config parse error
  ([6037d5c](https://codeberg.org/alinnow/searchix/commit/6037d5c33e9b98360f61a4fe4f09659d0e3e6160))

### [v0.1.3](https://codeberg.org/alinnow/searchix/compare/v0.1.2...v0.1.3) (2024-05-17)

#### Features

- **ui:** enable resetting form by clicking current source link
  ([6ea953b](https://codeberg.org/alinnow/searchix/commit/6ea953b1c9f9c352a9619e42a6f0ce80d183d7f8))
- render markdown in packages' LongDescription
  ([934a069](https://codeberg.org/alinnow/searchix/commit/934a0693213df901078cd1bc75a5b7e1b5c87555))
- use attribute instead of name for packages
  ([698881a](https://codeberg.org/alinnow/searchix/commit/698881aa577d2c209f162db9263fab17ced3643f))

### [v0.1.2](https://codeberg.org/alinnow/searchix/compare/v0.1.1...v0.1.2) (2024-05-17)

#### Fixes

- **css:** pagination styles not applied
  ([f10543d](https://codeberg.org/alinnow/searchix/commit/f10543d0aff0e203dd57287dc2b736c686e69102))

### [v0.1.1](https://codeberg.org/alinnow/searchix/compare/v0.1.0...v0.1.1) (2024-05-17)

#### Fixes

- **nix:** missing version variable in build
  ([8a1c2e8](https://codeberg.org/alinnow/searchix/commit/8a1c2e805b70ea4c4730fe7d785b942f46a54dea))

## v0.1.0 (2024-05-17)

### Features

- display more information about packages
  ([8df293c](https://codeberg.org/alinnow/searchix/commit/8df293c0ad8b8be9f43f2d5dc1334b8546d6f470))
- error on non-matching configuration entries
  ([6488132](https://codeberg.org/alinnow/searchix/commit/64881323e89d6d9b96865a5fa6f2150b826683db))
- automatically add/remove sources to/from index
  ([9f241bf](https://codeberg.org/alinnow/searchix/commit/9f241bff227608dd53a250d012116077dce6dab6))
- support searching packages
  ([05e6197](https://codeberg.org/alinnow/searchix/commit/05e61978906a08132c4340a5f9ae518134dd0fa9))
- **index:** enable storage of package data
  ([42611df](https://codeberg.org/alinnow/searchix/commit/42611df8133fc88bac5947a65a18fa095d68a951))
- improve accessibility and browser behaviour
  ([054d2b3](https://codeberg.org/alinnow/searchix/commit/054d2b3bf7eb4c5a5401565e86031f22ea754062))
- exit after rebuilding index if `--replace` is specified
  ([d6e117b](https://codeberg.org/alinnow/searchix/commit/d6e117bb9aeb2bae23dedd51b59b6e6f139e8a4f))
- decode brotli-compressed files
  ([b8da487](https://codeberg.org/alinnow/searchix/commit/b8da487f774aa2399b79bbde5f2e6e13be4fce1c))
- extract and store package set (e.g. nodePackages)
  ([8b78a98](https://codeberg.org/alinnow/searchix/commit/8b78a98af573a77a6d6aa6a0b231d4b657e35726))
- **importer:** process packages from nixpkgs package.json
  ([51f774a](https://codeberg.org/alinnow/searchix/commit/51f774aaced998ad3c7c31e370aea22cc2fdbbc5))
- add caching header (except on error pages)
  ([eb88ab1](https://codeberg.org/alinnow/searchix/commit/eb88ab15adaf23a61c783afe76b646855814d38a))
- monitor import jobs via sentry
  ([fa97ca8](https://codeberg.org/alinnow/searchix/commit/fa97ca832d6dc301dabb5d9c76730630f99534e0))
- add browser search engines via opensearch description
  ([5521173](https://codeberg.org/alinnow/searchix/commit/5521173ea0e05bda93096b51f67c5e813cb1d87e))
- print clickable URL on startup in dev mode
  ([121ec2b](https://codeberg.org/alinnow/searchix/commit/121ec2b445636cd3acb59391e1b241abfe2954a5))
- render HTML error pages
  ([2d5ee7b](https://codeberg.org/alinnow/searchix/commit/2d5ee7b69cbe27e7e1abef7bf3451cb6455c4387))
- **module:** add default for Web.Environment
  ([634534d](https://codeberg.org/alinnow/searchix/commit/634534d3d22d8f458a70a1891d061c3cdefd4dd1))
- add DownloadOptions importer (fetches pre-built options.json)
  ([df3b6bf](https://codeberg.org/alinnow/searchix/commit/df3b6bfd571c38aa7099838aa43bb3aff0eff2bf))
- log on startup if index doesn't exist
  ([5f9c7ef](https://codeberg.org/alinnow/searchix/commit/5f9c7ef00b1ed56fb1abbade4d5dbc3041acd4fc))
- rely on system to show log times in production
  ([37afa6c](https://codeberg.org/alinnow/searchix/commit/37afa6ce04434b81fcb242f1160938f1cddbde1e))
- allow adding extra HTML to the body via configuration
  ([7077a37](https://codeberg.org/alinnow/searchix/commit/7077a3748fa545e8dee83d4f3464c55b4b459183))
- enable using channel URLs if unable to resolve via NIX_PATH
  ([fe79b45](https://codeberg.org/alinnow/searchix/commit/fe79b4509dd580542fa28186673826df603677bb))
- nix module
  ([fb25a97](https://codeberg.org/alinnow/searchix/commit/fb25a97b6c55c6d51e59d983fdb74b4917c97bf0))
- clicking another source re-uses query
  ([42bcd10](https://codeberg.org/alinnow/searchix/commit/42bcd10f010f884bddb0142c31b526faace24186))
- **js:** save and restore input state with history
  ([b517b1e](https://codeberg.org/alinnow/searchix/commit/b517b1e0e1933b310bcbbd718fc664f87226d3b6))
- version search index mapping schema and warn if outdated
  ([76f18d6](https://codeberg.org/alinnow/searchix/commit/76f18d677b9b0bb70f701824e715f2ed10aeb8a0))
- dynamically link to sources
  ([631c02a](https://codeberg.org/alinnow/searchix/commit/631c02ac689b3366987918737e2b09f52fb16dda))
- format server listen address as URL
  ([f42dfe7](https://codeberg.org/alinnow/searchix/commit/f42dfe7b5fd3d60c9558eb520d33c7c8fd1ba201))
- enable index path to be configured
  ([0986e86](https://codeberg.org/alinnow/searchix/commit/0986e8699e87addccbef43c0ce30a466988fed15))
- **importer:** make timeout configurable via command-line
  ([1db7c2a](https://codeberg.org/alinnow/searchix/commit/1db7c2a1630c3d80c3c696c8fabaa5476eb2184e))
- **importer:** add --replace flag to overwrite existing index
  ([778e1d7](https://codeberg.org/alinnow/searchix/commit/778e1d774f574f45c179d0fecf0c1da9cc359b80))
- allow configuration of log level
  ([99d5507](https://codeberg.org/alinnow/searchix/commit/99d5507b223baa6c780707d32c7bbbb8611034f0))
- enable setting config file path via command line
  ([14c1dda](https://codeberg.org/alinnow/searchix/commit/14c1ddaf7c5468c5966b0c1585f19a5ff39c206d))
- import sources from configuration in go code and index options
  ([e062ca7](https://codeberg.org/alinnow/searchix/commit/e062ca72b222b890e345548bd8422d5df98e9fef))
- make configuration optional
  ([967f6fd](https://codeberg.org/alinnow/searchix/commit/967f6fdf5c1693d3aa27079b3ae28768fb7356c6))
- improve search by using a custom tokeniser for option names
  ([24b81a6](https://codeberg.org/alinnow/searchix/commit/24b81a63620667cd9c9f50952bbe881c08f16048))
- paginate search results
  ([973345a](https://codeberg.org/alinnow/searchix/commit/973345ad50f9b237714fcb364cf7f665b3909f9d))
- search multiple sources
  ([212e5cf](https://codeberg.org/alinnow/searchix/commit/212e5cf6621c99e46dbb37c860dab8938968bb19))
- search one set of options
  ([c15b142](https://codeberg.org/alinnow/searchix/commit/c15b142b18dcdc7f5ab6d5f1afca8ae1696692cc))
- link to exact commits in NixOS/nixpkgs
  ([8c13320](https://codeberg.org/alinnow/searchix/commit/8c1332020d25f74baa463bb1cec0e6783f565034))
- render markdown examples
  ([c0fbf11](https://codeberg.org/alinnow/searchix/commit/c0fbf11f843af84e8891a708c4d217dd6c523473))
- link to nixpkgs on github for declarations
  ([71de547](https://codeberg.org/alinnow/searchix/commit/71de5479bc625aa101f72a24b056b687b0630a9a))
- auto-linkify URLs in markdown
  ([0d65324](https://codeberg.org/alinnow/searchix/commit/0d653242b21abd1678df35e2ba6c30cc23733eb1))
- render markdown in option descriptions
  ([2430f46](https://codeberg.org/alinnow/searchix/commit/2430f46a9948b38b06880606a95dec357d01f619))
- browser history management
  ([87ec9ec](https://codeberg.org/alinnow/searchix/commit/87ec9ecf21781c5289257750fd41c6b9991f1f6e))
- show metadata in grid
  ([b682ed9](https://codeberg.org/alinnow/searchix/commit/b682ed936d8185e91331e9f66fc0e13072f8b545))
- frontend search implementation
  ([5b9e67f](https://codeberg.org/alinnow/searchix/commit/5b9e67fd5129dec75169a1a070c70f910dff6da2))
- create nixos options.json importer
  ([1fb64b0](https://codeberg.org/alinnow/searchix/commit/1fb64b09e2dba34784732a61cbebc119cd847803))
- add basic input frontend
  ([0945da8](https://codeberg.org/alinnow/searchix/commit/0945da8f9654252008ff6bd56022f9363fd3cca6))
- add css
  ([177c2de](https://codeberg.org/alinnow/searchix/commit/177c2de3f7c2b1cee8dbea4acaa896a334ba3162))
- watch and live reload in development
  ([8411fb4](https://codeberg.org/alinnow/searchix/commit/8411fb4a3acebe46acaf7a2ff6c4e58018737d65))
- serve a very basic html template
  ([7360307](https://codeberg.org/alinnow/searchix/commit/73603079e29bc89c54296a9e12b5a779cd84c023))

### Fixes

- render errors as HTML given correct content-type
  ([2edf3f0](https://codeberg.org/alinnow/searchix/commit/2edf3f098dd1bb79fd97a8f30fe28938cbde9f3c))
- missing last page of search results
  ([f9b0340](https://codeberg.org/alinnow/searchix/commit/f9b034027565bbc59b36d1a318716088c3a77977))
- store only data from current entity in index
  ([3b877b3](https://codeberg.org/alinnow/searchix/commit/3b877b39b424ded0f776d1dc71d07616a47bfbd4))
- nil pointer when updating all sources
  ([041ebd2](https://codeberg.org/alinnow/searchix/commit/041ebd2291db60e73d7fc213cd8a74d33b7f089b))
- don't panic when re-indexing if fetcher did not update
  ([43fe431](https://codeberg.org/alinnow/searchix/commit/43fe431c3aff36d52a630ca670f9fc4ee161962f))
- **css:** support browsers without css nesting
  ([29f2099](https://codeberg.org/alinnow/searchix/commit/29f2099e68a3226060308c66e5b1aa19fb2ba8e0))
- **logging:** track whether headers are written properly
  ([c5c9f72](https://codeberg.org/alinnow/searchix/commit/c5c9f726cae4cf4488312338da9599da53ac8b6c))
- rebuild index when --replace is specified
  ([d558039](https://codeberg.org/alinnow/searchix/commit/d558039919b6198a246a6a3fd007276191cb4b2f))
- bad git conflict resolution
  ([d65d80e](https://codeberg.org/alinnow/searchix/commit/d65d80e1133eebf11cc1adb3438a1cc53635185f))
- **flake:** remove overlay export
  ([8e22448](https://codeberg.org/alinnow/searchix/commit/8e22448e4dd53b284ba48bdf4ebf810bec489b41))
- **importer:** don't create duplicate nested directories
  ([a94632e](https://codeberg.org/alinnow/searchix/commit/a94632eb788fd014bea0917a683896c44866f431))
- **css:** barely-visible text in dark mode
  ([3fd832b](https://codeberg.org/alinnow/searchix/commit/3fd832b8f86641ea6046a776dda8f54a17834d7b))
- **module:** include nix in path
  ([67e1a28](https://codeberg.org/alinnow/searchix/commit/67e1a28ad3972a31bb995eaeec83434221ed5cf6))
- **config:** show read URL on parse failure
  ([45432ae](https://codeberg.org/alinnow/searchix/commit/45432ae8006322e89f89f43423ff26c99d98de10))
- **module:** allow setting sentry DSN
  ([f700a32](https://codeberg.org/alinnow/searchix/commit/f700a3284ecc88bb2af1df5f0c8d242290a549cc))
- don't panic on TOML decode error
  ([53c9fb2](https://codeberg.org/alinnow/searchix/commit/53c9fb2f72cbc82177404b76a4aad14283f8ced1))
- linking to wrong declaration URLs
  ([df77930](https://codeberg.org/alinnow/searchix/commit/df77930bef039e4aef024574c60fb64d75287eee))
- **import:** don't remove index if there are errors
  ([ef50bbf](https://codeberg.org/alinnow/searchix/commit/ef50bbf934a3024fdb82c31f5dc38e017603a86f))
- **js:** make reload after error execute attempted search
  ([84acac6](https://codeberg.org/alinnow/searchix/commit/84acac6ddf46351d4c049c7bb25d72584b2c0fb7))
- js error when no results found
  ([628c2a8](https://codeberg.org/alinnow/searchix/commit/628c2a858191cde34478b760b3c8c6c23a420088))
- missing menu on search pages
  ([793c89a](https://codeberg.org/alinnow/searchix/commit/793c89a87cb58f09678bdf150694a369c088d914))
- no results for home-manager search
  ([0eccb18](https://codeberg.org/alinnow/searchix/commit/0eccb18ab0a7241c1f6f5fee99a3961101e35392))
- **config:** add default channel name for home-manager
  ([3554d8a](https://codeberg.org/alinnow/searchix/commit/3554d8afb600b2eb62fbe7c2d4792debb9ce7a60))
- missing menu on search page
  ([42dd2f6](https://codeberg.org/alinnow/searchix/commit/42dd2f64f288a7474f356a8ae90e7409b7e38fb6))
- pointer type mismatch
  ([8ab9adf](https://codeberg.org/alinnow/searchix/commit/8ab9adf719090e213bd91b2747986ba6ef94576d))
- **js:** pushstate also on errors, so that the user can refresh
  ([453ae85](https://codeberg.org/alinnow/searchix/commit/453ae8569ab58fcc4ad61c461adc4489b9443cf8))
- **js:** null value error for pagination
  ([83f6751](https://codeberg.org/alinnow/searchix/commit/83f6751139ed2dbf2f7c2f235e03d91d632e38bb))
- incorrect default/example values stored after processing
  ([9f1a4aa](https://codeberg.org/alinnow/searchix/commit/9f1a4aaa9960fe3ab664033e1c1f42e62c3fb2e8))
- incorrect data shown for default and example option keys
  ([5e008c2](https://codeberg.org/alinnow/searchix/commit/5e008c251cf0e91ca6cdf9d91faaaa4975636541))
- panic on importing darwin options
  ([9b24b80](https://codeberg.org/alinnow/searchix/commit/9b24b805b45c71bb54dacc1112380357050f48af))
- use options exposed via user's <nixpkgs> channel instead of ours
  ([1fdd1f8](https://codeberg.org/alinnow/searchix/commit/1fdd1f839782339f39a16ba3100dec129cd40a0c))
- **css:** first grid row too tall
  ([3c54a3f](https://codeberg.org/alinnow/searchix/commit/3c54a3f63ad9b9ccaf5ca064ec3fd730164aad4a))

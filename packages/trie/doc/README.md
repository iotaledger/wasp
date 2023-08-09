# Trie package

![`Trie class diagram`](https://www.plantuml.com/plantuml/png/TLDDRnen4BtlhvXoOYk1GY9HQGKegbgfKczfbGj1bS53ri9hH_QGf2NyzzfcrsiSSe9zyzvyZD_SMcA6zeqic9JwvKyZNeLwB0fBPhyX-6q4tY7ZQE1G02ZDyHTfWrN_ry56QwhW1xDrSOpII0XACg9J_hm_PNJeCvFx7CvIV6F4GeR3Lg3aHtXYL7z_9LHMQ5N1ShN-I-Whe6c4Oh82skYc4TIW8eTlQY6vGK-TJ5UXIO2UaVUTgaDTpeWbPOIzzqrNiDPQ9h8xt6xqNf4D8lj-9gK9dOX8Dw2tQPcsY4iDAAX6KpbaT5eE3CKM9AfX-2fX1jERCeHhrtQBkczV4urWKln33ip2iWwLttpcbOk-kBm89n3ci6pdWE44re9AU0jLFBk4uHUFsN9LeEBW6uzZ-cN1uVquxLwNrrTXKQ6xHFt4DZlsYC3NC9lv9rqpYuj5sDLMIz_JLVW0u6sqjuo3ZprlalEYYJBYTKxqTehFTCwzkPGq8xkXatFuF1hr5iy3VXjLE8iYUdWyNgHNCDZDUmSygETvHnn_TVkpmt9mBarFPU1DyQap_BXzol91xUR15J5oieVFVWubkGUJGMZP_r9w5ftY8hMTiETVOarRMZp18YuiZDH9AcYOSkwmMRnVV_mNTH_5idUxHSNtZVmF)

[Edit](https://www.plantuml.com/plantuml/uml/TLDDRnen4BtlhvXoOYk1GY9HQGKegbgfKczfbGj1bS53ri9hH_QGf2NyzzfcrsiSSe9zyzvyZD_SMcA6zeqic9JwvKyZNeLwB0fBPhyX-6q4tY7ZQE1G02ZDyHTfWrN_ry56QwhW1xDrSOpII0XACg9J_hm_PNJeCvFx7CvIV6F4GeR3Lg3aHtXYL7z_9LHMQ5N1ShN-I-Whe6c4Oh82skYc4TIW8eTlQY6vGK-TJ5UXIO2UaVUTgaDTpeWbPOIzzqrNiDPQ9h8xt6xqNf4D8lj-9gK9dOX8Dw2tQPcsY4iDAAX6KpbaT5eE3CKM9AfX-2fX1jERCeHhrtQBkczV4urWKln33ip2iWwLttpcbOk-kBm89n3ci6pdWE44re9AU0jLFBk4uHUFsN9LeEBW6uzZ-cN1uVquxLwNrrTXKQ6xHFt4DZlsYC3NC9lv9rqpYuj5sDLMIz_JLVW0u6sqjuo3ZprlalEYYJBYTKxqTehFTCwzkPGq8xkXatFuF1hr5iy3VXjLE8iYUdWyNgHNCDZDUmSygETvHnn_TVkpmt9mBarFPU1DyQap_BXzol91xUR15J5oieVFVWubkGUJGMZP_r9w5ftY8hMTiETVOarRMZp18YuiZDH9AcYOSkwmMRnVV_mNTH_5idUxHSNtZVmF)


## Example

`TestBasic` in `packages/trie/test/trie_test.go` generates 4 trie roots:

```
trie root 1: (empty trie)
trie root 2:
  set 0x61 "a" = "a"
  set 0x62 "b" = "b"
trie root 3:
  set 0x62 "b" = "bb"
trie root 4:
  del 0x61 "a"
  set 0x636363646464 "cccddd" = "c"
  set 0x636363656565 "ccceee" = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
```

The resulting trie has the following structure:

```
[trie store]
├─ [] c:534f98b3ad630819d284287b647283a1d5dbcf90 ext:[] term:<nil>
├─ [] c:7ec331767219528ab3c9e864ad9422e9f831ec5e ext:[] term:<nil>
│  └─ [6] c:a00e505e10971ec248b1404ef9dcc6c6c493a6c9 ext:[] term:<nil>
│     ├─ [1] c:81db21106a17dd57e6099402bbe5543a015193d0 ext:[] term:"a"
│     └─ [2] c:b23756724eca8e6197bb6b6cbfc9725067b36d9c ext:[] term:"b"
├─ [] c:27806fe716c7f11add3ada0a356c0fb9c9377c5a ext:[] term:<nil>
│  └─ [6] c:fd3f071332a7fcca857cf49df735ff49c4539d07 ext:[] term:<nil>
│     ├─ [1] c:81db21106a17dd57e6099402bbe5543a015193d0 ext:[] term:"a"
│     └─ [2] c:bae0c3296e8fa86c0d6f6621200f987cc01a6c0a ext:[] term:"bb"
└─ [] c:b8cc8cb105beb3ee7100049b91d3d8b0c49ba05a ext:[] term:<nil>
   └─ [6] c:65a4c9c112d3e8e995c36915c527f4b85a9507fb ext:[] term:<nil>
      ├─ [2] c:bae0c3296e8fa86c0d6f6621200f987cc01a6c0a ext:[] term:"bb"
      └─ [3] c:9af38ad122bdf3ea08cf146f0f8608ae00dc7624 ext:[6 3 6 3 6] term:<nil>
         ├─ [4] c:fe360246e69933a23be31ae26653c0fd68d790a8 ext:[6 4 6 4] term:"c"
         └─ [5] c:a2185f866895f47d830bb8aa096a9a51e6d9a25b ext:[6 5 6 5] term:d25eb7361e92d86c9fcf3f7f602217fc45d86290

[value store]
   d25eb7361e92d86c9fcf3f7f602217fc45d86290: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

[node refcounts]
   65a4c9c112d3e8e995c36915c527f4b85a9507fb: 1
   534f98b3ad630819d284287b647283a1d5dbcf90: 1
   a2185f866895f47d830bb8aa096a9a51e6d9a25b: 1
   fe360246e69933a23be31ae26653c0fd68d790a8: 1
   27806fe716c7f11add3ada0a356c0fb9c9377c5a: 1
   9af38ad122bdf3ea08cf146f0f8608ae00dc7624: 1
   b23756724eca8e6197bb6b6cbfc9725067b36d9c: 1
   7ec331767219528ab3c9e864ad9422e9f831ec5e: 1
   bae0c3296e8fa86c0d6f6621200f987cc01a6c0a: 2
   fd3f071332a7fcca857cf49df735ff49c4539d07: 1
   81db21106a17dd57e6099402bbe5543a015193d0: 2
   a00e505e10971ec248b1404ef9dcc6c6c493a6c9: 1
   b8cc8cb105beb3ee7100049b91d3d8b0c49ba05a: 1

[value refcounts]
   d25eb7361e92d86c9fcf3f7f602217fc45d86290: 1
```

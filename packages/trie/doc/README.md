# Trie package

![Trie class diagram](https://www.plantuml.com/plantuml/png/RPDHQzj03CVV_IcqFCM3XAQ538uabB71OB27KFgIXd7soV6HyrsuKzfkQRxxJZVNSKsUafD-alwIhFj0qjEnCPaqnhtyRlIhjhii2-IF9S63rxY6KmOGo7aKyFDnZLAXpyuTyyif_2P99knKQJJg30T9zVTNGXkqim8ss-8FOVGJ-aPRQGGybAvfD3LeAThBMXlbaT7PVFj3XfgD7I4WcRgY_UXKLqK1lJ8wF4fpH065SuR3h2C4SWTmrQ0oHqOUK5EDNaJB19uv6CQw0eDsd5uvB61-boTt7RMqrlZcuQ3NqkY_g406sDHfH6-SUHkzBp4lmdyZB67JYeFn30fDERobAapdCFx2jf1lCxGAM7NNCay0Jk-z6Ao8csb7jRczIRqdRRtOWoptSeUW2bWiU4k53BXQfPbLph2Yc_G0KAqNo6lafOPeFb-usJMEEnc2jqFGncHhh4ftj3HWxuOiZtK0nvQxRsmwNLtq2HZjbvTbFjddfMt1wTSuZ33EHnz21qhw7n5xkuOuCs__Kxkg_CdkbxsjSwNqRpJCCTJVLH_DHl1vCac_lHZHNSHrbDShuC9Ve-OTrXvHbiusoNd71JJIoXs6x1wjYlV-3m00)

[Edit](https://www.plantuml.com/plantuml/uml/RPDHQzj03CVV_IcqFCM3XAQ538uabB71OB27KFgIXd7soV6HyrsuKzfkQRxxJZVNSKsUafD-alwIhFj0qjEnCPaqnhtyRlIhjhii2-IF9S63rxY6KmOGo7aKyFDnZLAXpyuTyyif_2P99knKQJJg30T9zVTNGXkqim8ss-8FOVGJ-aPRQGGybAvfD3LeAThBMXlbaT7PVFj3XfgD7I4WcRgY_UXKLqK1lJ8wF4fpH065SuR3h2C4SWTmrQ0oHqOUK5EDNaJB19uv6CQw0eDsd5uvB61-boTt7RMqrlZcuQ3NqkY_g406sDHfH6-SUHkzBp4lmdyZB67JYeFn30fDERobAapdCFx2jf1lCxGAM7NNCay0Jk-z6Ao8csb7jRczIRqdRRtOWoptSeUW2bWiU4k53BXQfPbLph2Yc_G0KAqNo6lafOPeFb-usJMEEnc2jqFGncHhh4ftj3HWxuOiZtK0nvQxRsmwNLtq2HZjbvTbFjddfMt1wTSuZ33EHnz21qhw7n5xkuOuCs__Kxkg_CdkbxsjSwNqRpJCCTJVLH_DHl1vCac_lHZHNSHrbDShuC9Ve-OTrXvHbiusoNd71JJIoXs6x1wjYlV-3m00)

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

![Trie nodes](https://www.plantuml.com/plantuml/png/ZPJBRhCm48NtFCM8fQL4WBdmH_gZF4EtLHV6DbLKBX4nLQGexrwEhBeOahJEYYypXpFZ6Gvd8mOpjarW-XEPdghUcBubTHoqDCU-2uAJa1_YNTEdQ4Pzfcj0qsd5YLIIQVk8O_-d_tiAPKYCg8W2u5jm-z7eUkL9P3ojyKY8SxHAU34JPipeeJrsec4C-eo5LSYOpN99PQogvqbYdIIYQ5v2J6opsHEa74iiWwT5Sg4bdoNdDX2IguA4Zdufw4v5jUBrtyn1Vv6S56MqDWV9fRbQytmMgIz6T-Vfvk7iDK_pz8r2fS9dxkMTpYPn-LWroxD9LUjdgXWl9-jq_ucQFzRbqsyIbZBSRgDBxi9IMpvATIop34ONtlDCf7Hzs6GSRXdSB9AhKifruo6vNeT1rx8VEpDV2Qp8ouBZu49HPEm8biSdU8jFCAC-mVo098wW23y6C-mJTCadM0YVe7AVe6N-kp20VWpkn97vJUDxShSJ68nvnfG3UxlTzmGHEeweuA0xDRRD5m00)

[Edit](https://www.plantuml.com/plantuml/uml/ZPJDRgim48NtFCM8fQL4WFdXHhfHdg7RgWjZcofgW8WuLQGexrwEhBhO96cT5Lzc3cV6C-nEHepct1qYCnp93DGCnWmTgVsefTaFVHGTXVmGxve-nU6iJtIDQ3gTE9BA2cqVSUp_z7zBiOo94LL917oBRZyQ_Q0yYOBdBU4HSOus6QzcmemPFNRxNXmOWpvZSHCpPhDSKfsgvdcIM2Q999cNa19RhRO4ASSoo-0fIAwqvBFaqWP2Sb6GqCal1Tsfg5My_Pivw0V9MLKdItEG3CsrcfUpI7qnkZzDDmzdhsLUvcyKAnK_So_lT3PBwyUgHLbCgbe_KyLuELcd-q-q_B4kdtwJi9JnUfrSIXUMs_9HgcE5PJ0QU2yBaTFrOPDnk6Lm4oGtjPJBnaDokGwZRkLmxiry9R0YBm77mOMYoDGHB8zFy1QVO4PzWla1oHn14NuCPjWdw99Fi10-GUK-GSlyTs40lHWxVY3XJUDxShSJ68nvnfG3UvlDzmGHEeweQATxXO_R5m00)

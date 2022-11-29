# Trie package

![`Trie class diagram`](https://www.plantuml.com/plantuml/png/RL9DZzem4BtxLqnpIYe5bLQYKWHeLRLIfTxQQYyi1mSF6ml7Njd3zWFuxsixE1Z12VBclV6RcVVWY5lQzueviliDmMyhyIToWHOE340RWR_8M6mkVpriZQ46ldFNiHDBqf4GbMHbKlvu73fwz9Mh_GrytU8h9nux7BOIbJZ12wVksrz2xQJH3QpMxJ_2y0BQNcgk6g2DwNj9FMho-AQJIbWCrEbi7Kq2Z8mRtxawlYkyWUmPwHw3wGPQOrIGQKC8LZvt16QRgyzQhm2KrA5jF58FCqCfjw1Gb_6hWZdCFbMnt7at0ng-6O13AxcI_r40Tx3gufAEeVFQL__ulWW320jOdUr1EOLMKWN7-4fWLr1-3fYhrWorWE0x3Hrt08URQSxRMdty4CUFvZBn2z_i-3E2Q64-3uTgkSFbCgujrIudlkMSCbuAo5sQDvObyNrTP_6xBaJBJKma6-CpcIpp01QxnULAJ_fraOYJBtv8LrR5jJHFQH4EzovbRN9UT_MaTujukR4od31qluQotiMq29RZB-M9J8hxr6722_yUQvPeAVriN5WSAKaQwBdswtTPtTJr4aJBs0DgiU_L6m00)

[Edit](https://www.plantuml.com/plantuml/uml/RL9DZzem4BtxLqnpIYe5bLQYKWHeLRLIfTxQQYyi1mSF6ml7Njd3zWFuxsixE1Z12VBclV6RcVVWY5lQzueviliDmMyhyIToWHOE340RWR_8M6mkVpriZQ46ldFNiHDBqf4GbMHbKlvu73fwz9Mh_GrytU8h9nux7BOIbJZ12wVksrz2xQJH3QpMxJ_2y0BQNcgk6g2DwNj9FMho-AQJIbWCrEbi7Kq2Z8mRtxawlYkyWUmPwHw3wGPQOrIGQKC8LZvt16QRgyzQhm2KrA5jF58FCqCfjw1Gb_6hWZdCFbMnt7at0ng-6O13AxcI_r40Tx3gufAEeVFQL__ulWW320jOdUr1EOLMKWN7-4fWLr1-3fYhrWorWE0x3Hrt08URQSxRMdty4CUFvZBn2z_i-3E2Q64-3uTgkSFbCgujrIudlkMSCbuAo5sQDvObyNrTP_6xBaJBJKma6-CpcIpp01QxnULAJ_fraOYJBtv8LrR5jJHFQH4EzovbRN9UT_MaTujukR4od31qluQotiMq29RZB-M9J8hxr6722_yUQvPeAVriN5WSAKaQwBdswtTPtTJr4aJBs0DgiU_L6m00)


## Example

Given the following pairs of key-values:

```
"b"      => "bb"
"cccddd" => "c"
"ccceee" => "c" * 70
```

the trie will have the following structure:

```
deb052a7efca855ad373283fb1fcf73cdb44667f
  Key:  "" ()
  Full key: ""
  child(6): 6174a77493e10eae75d828ffb0fddd7dc9adfa10
6174a77493e10eae75d828ffb0fddd7dc9adfa10
  Key:  "`" (0x60)
  Full key: "`" (0x60)
  child(2): 7e84c82d6daa06e959d1d9c43c1de6c00c3ab456
  child(3): a3acfd855f84c4b665324397b4583947f19595f7
7e84c82d6daa06e959d1d9c43c1de6c00c3ab456
  Key:  "b" (0x62)
  Full key: "b"
  Terminal:    "bb"
a3acfd855f84c4b665324397b4583947f19595f7
  Key:  "c" (0x63)
  Extension: 0x63636
  Full key: "ccc`"
  child(4): a3c2a38f7da19d791df3fda09dfde7f60b5e7d7f
  child(5): c4cf202df0ec08a0cc5a5d0c7bfe972433d25253
a3c2a38f7da19d791df3fda09dfde7f60b5e7d7f
  Key:  "cccd" (0x63636364)
  Extension: "dd" (0x6464)
  Full key: "cccddd"
  Terminal:    "c"
c4cf202df0ec08a0cc5a5d0c7bfe972433d25253
  Key:  "ccce" (0x63636365)
  Extension: "ee" (0x6565)
  Full key: "ccceee"
  Terminal: 0x14d25eb7361e92d86c9fcf3f7f602217fc45d86290
    (in valueStore: 0x14d25... => "c" * 70)
```

![Example](https://www.plantuml.com/plantuml/png/RP11ReCm44NtFiLeLacaI3GOqqsZ7ANsU2fGCOx0aL9LlNipmQA88dR_VxydvZ8cEwHRw9HAyGefWeifnG2-7PXI6tkbWhs2BUa4tmg0xudxyP73sncAkz560tgFZ-gJlt9OoYULb4Jafq7Y8RIzxMI5XEhdYIPD7vleobI0p5leQYg9Y9dNwCQEpGu9uG1G5_kiiQMbJOtm1FM-5kKYDzIqdT_fFHoEn_Fp7dDOvL3-9DjgVjLYMm81tRyhE8VvXNcfroy05-PoDbkQmrCEHPSvdbsnvE0VmanhKSbQNRZtjz3z0W00)

![Edit](https://www.plantuml.com/plantuml/uml/RP11ReCm44NtFiLeLacaI3GOqqsZ7ANsU2fGCOx0aL9LlNipmQA88dR_VxydvZ8cEwHRw9HAyGefWeifnG2-7PXI6tkbWhs2BUa4tmg0xudxyP73sncAkz560tgFZ-gJlt9OoYULb4Jafq7Y8RIzxMI5XEhdYIPD7vleobI0p5leQYg9Y9dNwCQEpGu9uG1G5_kiiQMbJOtm1FM-5kKYDzIqdT_fFHoEn_Fp7dDOvL3-9DjgVjLYMm81tRyhE8VvXNcfroy05-PoDbkQmrCEHPSvdbsnvE0VmanhKSbQNRZtjz3z0W00)

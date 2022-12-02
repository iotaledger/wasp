# Trie package

![`Trie class diagram`](https://www.plantuml.com/plantuml/png/RPDDZzem48Rl_XNJd99AeD9AL8e0hVfnwgr5N2W79nas5ewziWVwiV3VrpPn4fISYD-PUJmpVdcon0sT6yaicND9i3K2VoAlqC0h0r2wmb-a3URzfvEDrfJ3xgjk6a4DAc8t5WbTNW2G4H7mOtS7E_N2vNaaFoA-SUA1y7eKAAiRl72gl_ybL5PebSF-KFmZtmkXQOJYCeVje1-0CXvYftsa5h8ow26BvM5wYaZjWw6PYXCVt2tipa-IGw6rzDLA4u8HppM1Fav0MeGh11HKeycTChll0r9n18ag3QSCLh3yb8LmpMtTgkxC02mQmQNuZnLm2_iwBF3goGSN_UEu27SHrh1lkuHq0OMW3AxX5f-TGVhPQpOxC5GL-FQx5Lom96s62uW3IZpw-KOP363_yKDOJkXvm7pXvOa_8rG5h5R-RGjkSRC9IusntXNZd5kV5pqrdUPw77Zu_MmsCEvZODH-7dpDElVDvvkXVvFJki-TVt53sZJxhIxhM12rv14nD7lYt_XTzCMyVHb7EQt38tUHK1JeZzA0dImIFqRu7JijXsshJLUCX--_fogMaoNjFubrEmuJn_8SIeESq3BsXAewD_8D)

[Edit](https://www.plantuml.com/plantuml/uml/RPDDZzem48Rl_XNJd99AeD9AL8e0hVfnwgr5N2W79nas5ewziWVwiV3VrpPn4fISYD-PUJmpVdcon0sT6yaicND9i3K2VoAlqC0h0r2wmb-a3URzfvEDrfJ3xgjk6a4DAc8t5WbTNW2G4H7mOtS7E_N2vNaaFoA-SUA1y7eKAAiRl72gl_ybL5PebSF-KFmZtmkXQOJYCeVje1-0CXvYftsa5h8ow26BvM5wYaZjWw6PYXCVt2tipa-IGw6rzDLA4u8HppM1Fav0MeGh11HKeycTChll0r9n18ag3QSCLh3yb8LmpMtTgkxC02mQmQNuZnLm2_iwBF3goGSN_UEu27SHrh1lkuHq0OMW3AxX5f-TGVhPQpOxC5GL-FQx5Lom96s62uW3IZpw-KOP363_yKDOJkXvm7pXvOa_8rG5h5R-RGjkSRC9IusntXNZd5kV5pqrdUPw77Zu_MmsCEvZODH-7dpDElVDvvkXVvFJki-TVt53sZJxhIxhM12rv14nD7lYt_XTzCMyVHb7EQt38tUHK1JeZzA0dImIFqRu7JijXsshJLUCX--_fogMaoNjFubrEmuJn_8SIeESq3BsXAewD_8D)


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

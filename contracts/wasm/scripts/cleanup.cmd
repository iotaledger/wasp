cd %1
if not exist schema.yaml goto :xit
schema -go -rust -ts -clean
if exist ts\%1\tsconfig.json del ts\%1\tsconfig.json
:xit
cd ..

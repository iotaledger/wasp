cd %1
schema -go -rust -ts -clean
del ts\%1\tsconfig.json
cd ..

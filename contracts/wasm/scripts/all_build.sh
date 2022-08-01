#!/bin/bash
bash core_build.sh
bash go_all.sh -force
bash ts_all.sh -force
bash rust_all.sh -force
bash update_hardcoded.sh

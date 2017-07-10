#!/bin/bash
rm -rf assets/dist
npm run build
rice embed-go

# TODO 1. install golang, pnpm, nodejs

# TODO 2. build normally
# frontend
cd frontend
pnpm install
pnpm run build
# backend
cd ../
go mod download
go build

# TODO 3. copy output to desired location
# TODO 4. clean up

# Production Config Directory

Place your `config.yaml` here before running `make prod-up`.

```bash
cp ../../configs/config.example.yaml config.yaml
# Edit config.yaml with your production values
```

This directory is mounted into the backend container at `/app/configs/`.

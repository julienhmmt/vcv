# VCV FRONTEND

## Commandes pour le développement

### Installation d'Astro

```bash
$ cd ~/git/vcv

$ docker run -it --rm \
  -v $(pwd):/workdir \
  -w /workdir \
  node:25.1-alpine \
  ash

# dans le conteneur
npm create astro@latest -- --template minimal --add svelte
```

### Mise à jour du frontend

```bash
docker run -it --rm \
  -v $(pwd)/frontend:/workdir \
  -w /workdir \
  node:25.1-alpine \
  sh -c "npx @astrojs/upgrade"
```

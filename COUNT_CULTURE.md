# Our changes

Our initial changes can be viewed here

https://github.com/grafana/grafana/compare/v12.0.0...countculture-ai:grafana:features/v12.0.0-enhancements

These enhancements are little tweaks.

## TODO 

* Embed datebar panel?

## Syncing Fork

See https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork

## Creating a release branch

1. Create a branch from the release version e.g  `git checkout -b feature/v12.0.2-enhancements V12.0.2`
2. Then apply changes below 
3. Commit branch and create PR.

## Sources
* 
* public/app/core/components/AppChrome/MegaMenu/MegaMenu.tsx - if not admin/editor remove some items
* public/app/features/dashboard-scene/inspect/InspectDataTab.tsx - default downloadForExcel to true 
* public/app/features/inspector/InspectDataTab.tsx - default transforms on 
* public/app/features/dashboard-scene/inspect/PanelInspectDrawer.tsx - if not admin/editor remove all menu options other than inspect 
* public/app/features/dashboard-scene/scene/NavToolbarActions.ts - if not admin/editor remove share 
* public/app/features/dashboard-scene/scene/PanelMenuBehavior.tsx -- add quick link to download CSV

## Developing

To handle different `npm` versions install [NVM](https://github.com/nvm-sh/nvm)

```bash
nvm install lts/jod
nvm use lts/jod
nvm current
npm install -g yarn
```

See [dev guide](contribute/developer-guide.md)

Open two terminals

### Front end
```shell
yarn install --immutable
yarn start
```
### Back end
```shell
make run
```



## Building

```shell
make cc-docker
```

> Can take some time :-)

## Running








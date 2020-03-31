const importID = 'import-to-creamy-videos';
const tagImportID = 'import-to-creamy-videos-tagged';
const promptImportID = 'import-to-creamy-videos-prompt';

let tagGroups = [];
let tagSubmenus = {};

function syncMenus() {
  browser.contextMenus.removeAll();

  tagGroups = [];
  tagSubmenus = {};

  browser.contextMenus.create({
    id: importID,
    title: 'Import to Creamy Videos',
    contexts: ['all'],
  });

  browser.contextMenus.create({
    id: tagImportID,
    parentId: importID,
    title: 'Default',
    contexts: ['all'],
  });

  browser.contextMenus.create({
    id: promptImportID,
    parentId: importID,
    title: 'Ask',
    contexts: ['all'],
  });

  browser.storage.sync.get('tagGroups').then((settingItem) => {
    tagGroups = settingItem.tagGroups || [];

    tagGroups.forEach((submenu, i) => {
      const id = `${tagImportID}-${i}`;
      tagSubmenus[id] = i;
      browser.contextMenus.create({
        id,
        parentId: importID,
        title: submenu.text,
        contexts: ['all'],
      });
    });
  });
}

syncMenus();
if (!browser.storage.onChanged.hasListener(syncMenus)) {
  browser.storage.onChanged.addListener(syncMenus);
}

browser.notifications.onClicked.addListener(function () {
  browser.storage.sync.get('url').then((settingItem) => {
    const url = settingItem.url || 'http://localhost:4000/';
    browser.tabs.create({
      active: true,
      url: url,
    });
  });
});

browser.contextMenus.onClicked.addListener((info, tab) => {
  let tags;
  let tagPromise = null;

  if (info.menuItemId === importID || info.menuItemId === tagImportID) {
    tags = [];
  } else if (info.menuItemId === promptImportID) {
    tagPromise = browser.tabs.executeScript(tab.id, {
      code: `prompt('Enter tags, separated by comma:')`,
    }).then((promptResult, err) => {
      if (err) {
        return [];
      }

      tags = promptResult[0] || '';
      tags = tags.split(',');
      tags = tags.map(tag => tag.trim());
      return tags;
    });
  } else if (tagSubmenus[info.menuItemId] !== undefined) {
    const tagGroup = tagGroups[tagSubmenus[info.menuItemId]];
    tags = tagGroup.tags;
  } else {
    // unhandled menu item, ignore it
    console.warn('unhandled menu item', info.menuItemId, info);
    return;
  }

  if (!tagPromise) {
    tagPromise = Promise.resolve(tags);
  }

  tagPromise.then((tags) => {
    const url = info.linkUrl || tab.url;
    
    browser.storage.sync.get('url').then((settingItem) => {
      fetch(settingItem.url || 'http://localhost:4000/', {
        method: 'post',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'url=' + encodeURIComponent(url) + '&tags=' + encodeURIComponent(tags.join(',')),
      }).then(function (resp) {
        if (!resp.ok) {
          throw new Error(resp.statusText);
        }
        return resp.text();
      }).then(function () {
        browser.notifications.create({
          'type': 'basic',
          'title': 'Creamy Videos Importer',
          'message': 'Import queued successully!',
        });
      }).catch(function (ex) {
        browser.notifications.create({
          'type': 'basic',
          'title': 'Creamy Videos Importer',
          'message': 'Error importing video!\n' + ex.message,
        });
        throw ex;
      });
    });
  });
});

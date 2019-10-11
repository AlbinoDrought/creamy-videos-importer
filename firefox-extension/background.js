const importID = 'import-to-creamy-videos';

browser.contextMenus.create({
  id: importID,
  title: 'Import to Creamy Videos',
  contexts: ['all'],
});

browser.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId !== importID) {
    return;
  }

  const url = info.linkUrl || tab.url;

  browser.storage.sync.get('url').then((settingItem) => {
    fetch(settingItem.url || 'http://localhost:4000/', {
      method: 'post',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: 'url=' + encodeURIComponent(url),
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

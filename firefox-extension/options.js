function tagGroupsToText(tagGroups) {
  return tagGroups.map((group) => {
    return [group.text, ...group.tags].join(',');
  }).join('\n');
}

function textToTagGroups(text) {
  return text.split('\n').map((line) => {
    const columns = line.split(',');
    return {
      text: columns[0],
      tags: columns.slice(1),
    };
  }).filter(group => group.tags.length > 0);
}

function saveOptions(e) {
  e.preventDefault();
  browser.storage.sync.set({
    url: document.querySelector('#url').value,
    tagGroups: textToTagGroups(document.querySelector('#tag-groups').value || ''),
  });

  var submitButton = document.querySelector('button[type="submit"]');
  submitButton.innerText = 'Saved!';
  setTimeout(function () {
    submitButton.innerText = 'Save';
  }, 2500);
}

function restoreOptions() {
  function setCurrentChoice(result) {
    document.querySelector('#url').value = result.url || 'http://localhost:4000/';
    document.querySelector('#tag-groups').value = tagGroupsToText(result.tagGroups);
  }

  function onError(error) {
    console.log(`Error: ${error}`);
  }

  var getting = browser.storage.sync.get(['url', 'tagGroups']);
  getting.then(setCurrentChoice, onError);
}

document.addEventListener('DOMContentLoaded', restoreOptions);
document.querySelector('form').addEventListener('submit', saveOptions);

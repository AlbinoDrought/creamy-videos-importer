function saveOptions(e) {
  e.preventDefault();
  browser.storage.sync.set({
    url: document.querySelector('#url').value
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
  }

  function onError(error) {
    console.log(`Error: ${error}`);
  }

  var getting = browser.storage.sync.get('url');
  getting.then(setCurrentChoice, onError);
}

document.addEventListener('DOMContentLoaded', restoreOptions);
document.querySelector('form').addEventListener('submit', saveOptions);

import * as Base64 from "./base64.js";
import { chroma } from "./chroma.js";

document.addEventListener("DOMContentLoaded", () => {
  const darkMode =
    window.matchMedia?.("(prefers-color-scheme: dark)").matches;
  const style = document.createElement("style");
  const ref = document.querySelector("script");
  ref.parentNode.insertBefore(style, ref);

  const form = document.getElementById("chroma");
  const textArea = form.elements.text;
  const styleSelect = form.elements.style;
  const languageSelect = form.elements.language;
  const copyButton = form.elements.copy;
  const csrfToken = form.elements["gorilla.csrf.Token"].value;
  const output = document.getElementById("output");
  const htmlCheckbox = document.getElementById("html");

  (document.querySelectorAll(".notification .delete") || []).forEach((el) => {
    const notification = el.parentNode;
    el.addEventListener("click", () => {
      notification.parentNode.removeChild(notification);
    });
  });

  async function renderServer(formData) {
    return (
      await fetch("api/render", {
        method: "POST",
        mode: "cors",
        cache: "no-cache",
        credentials: "same-origin",
        headers: {
          "X-CSRF-Token": csrfToken,
          "Content-Type": "application/json",
        },
        redirect: "follow",
        referrer: "no-referrer",
        body: JSON.stringify(formData),
      })
    ).json();
  }

  async function renderWasm(formData) {
    return await chroma.highlight(
      formData.text,
      formData.language,
      formData.style,
      formData.classes,
    );
  }

  async function render(formData) {
    return chroma !== null ? renderWasm(formData) : renderServer(formData);
  }

  // https://stackoverflow.com/a/37697925/7980
  function handleTab(e) {
    let after;
    let before;
    let end;
    let lastNewLine;
    let changeLength;
    let re;
    let replace;
    let selection;
    let start;
    let val;
    if (
      (e.charCode === 9 || e.keyCode === 9) &&
      !e.altKey &&
      !e.ctrlKey &&
      !e.metaKey
    ) {
      e.preventDefault();
      start = this.selectionStart;
      end = this.selectionEnd;
      val = this.value;
      before = val.substring(0, start);
      after = val.substring(end);
      replace = true;
      if (start !== end) {
        selection = val.substring(start, end);
        if (~selection.indexOf("\n")) {
          replace = false;
          changeLength = 0;
          lastNewLine = before.lastIndexOf("\n");
          if (!~lastNewLine) {
            selection = before + selection;
            changeLength = before.length;
            before = "";
          } else {
            selection = before.substring(lastNewLine) + selection;
            changeLength = before.length - lastNewLine;
            before = before.substring(0, lastNewLine);
          }
          if (e.shiftKey) {
            re = /(\n|^)(\t|[ ]{1,8})/g;
            if (selection.match(re)) {
              start--;
              changeLength--;
            }
            selection = selection.replace(re, "$1");
          } else {
            selection = selection.replace(/(\n|^)/g, "$1\t");
            start++;
            changeLength++;
          }
          this.value = before + selection + after;
          this.selectionStart = start;
          this.selectionEnd = start + selection.length - changeLength;
        }
      }
      if (replace && !e.shiftKey) {
        this.value = `${before}\t${after}`;
        this.selectionStart = this.selectionEnd = start + 1;
      }
    }
    debouncedEventHandler(e);
  }

  function debounce(func, wait, immediate) {
    let timeout;
    return function () {
      const args = arguments;
      const later = () => {
        timeout = null;
        if (!immediate) func.apply(this, args);
      };
      const callNow = immediate && !timeout;
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
      if (callNow) func.apply(this, args);
    };
  }

  function getFormJSON() {
    return {
      language: languageSelect.value,
      style: styleSelect.value,
      text: textArea.value,
      classes: htmlCheckbox.checked,
    };
  }

  async function update(event) {
    try {
      const formData = getFormJSON();
      const value = await render(formData);

      if (value.language) {
        languageSelect.value = value.language;
      }
      style.innerHTML = `#output { ${value.background}}`;
      if (htmlCheckbox.checked) {
        output.innerText = value.html;
      } else {
        output.innerHTML = value.html;
      }
    } catch (error) {
      console.error("Error highlighting code:", error);
      // Fallback: display plain text
      if (htmlCheckbox.checked) {
        output.innerText = textArea.value;
      } else {
        output.innerHTML = `<pre>${textArea.value}</pre>`;
      }
    }

    if (event) {
      event.preventDefault();
    }
  }

  function share(event) {
    let data = JSON.stringify(getFormJSON());
    data = Base64.encodeURI(data);
    location.hash = `#${data}`;
    try {
      navigator.clipboard.writeText(location.href);
    } catch (e) {
      console.log(e);
    }
    event.preventDefault();
  }

  if (location.hash) {
    let json = Base64.decode(location.hash.substring(1));
    json = JSON.parse(json);
    textArea.value = json.text;
    languageSelect.value = json.language;
    styleSelect.value = json.style;
    htmlCheckbox.checked = json.classes;
    update(new Event("change"));
  } else if (darkMode) {
    styleSelect.value = "monokai";
    update(new Event("change"));
  }

  const eventHandler = (event) => update(event);
  const debouncedEventHandler = debounce(
    eventHandler,
    chroma === null ? 250 : 100,
  );

  languageSelect.addEventListener("change", eventHandler);
  styleSelect.addEventListener("change", eventHandler);
  htmlCheckbox.addEventListener("change", eventHandler);
  copyButton.addEventListener("click", share);

  textArea.addEventListener("keydown", handleTab);
  textArea.addEventListener("change", debouncedEventHandler);
});

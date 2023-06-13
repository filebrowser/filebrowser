export default function getRule(rules) {
  for (let i = 0; i < rules.length; i++) {
    rules[i] = rules[i].toLowerCase();
  }

  let result = null;
  let find = Array.prototype.find;

  find.call(document.styleSheets, (styleSheet) => {
    result = find.call(styleSheet.cssRules, (cssRule) => {
      let found = false;

      if (cssRule instanceof window.CSSStyleRule) {
        for (let i = 0; i < rules.length; i++) {
          if (cssRule.selectorText.toLowerCase() === rules[i]) {
            found = true;
          }
        }
      }

      return found;
    });

    return result != null;
  });

  return result;
}

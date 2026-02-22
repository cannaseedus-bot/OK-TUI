export function tokenize(text) {
  return text.split("").map((x) => x.charCodeAt(0));
}

export function detokenize(tokens) {
  return tokens.map((x) => String.fromCharCode(x)).join("");
}

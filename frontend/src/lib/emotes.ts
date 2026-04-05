const EMOTES: [RegExp, string][] = [
  // Text faces
  [/:\)/g,  "😊"],
  [/:\(/g,  "😢"],
  [/:D/g,   "😄"],
  [/;D/g,   "😁"],
  [/;P/g,   "😜"],
  [/;-\)/g, "😉"],
  [/;\)/g,  "😉"],
  [/:P/g,   "😛"],
  [/:O/g,   "😮"],
  [/:o/g,   "😮"],
  [/:'/g,  "😭"],
  [/>:\(/g, "😠"],
  [/:\*/g,  "😘"],
  [/B\)/g,  "😎"],
  [/<3/g,   "❤️"],
  [/<\/3/g, "💔"],

  // Shortcodes
  [/:fire:/g,       "🔥"],
  [/:heart:/g,      "❤️"],
  [/:thumbsup:/g,   "👍"],
  [/:thumbsdown:/g, "👎"],
  [/:laugh:/g,      "😂"],
  [/:cry:/g,        "😭"],
  [/:clap:/g,       "👏"],
  [/:wave:/g,       "👋"],
  [/:ok:/g,         "👌"],
  [/:100:/g,        "💯"],
  [/:star:/g,       "⭐"],
  [/:poop:/g,       "💩"],
  [/:eyes:/g,       "👀"],
  [/:muscle:/g,     "💪"],
  [/:pray:/g,       "🙏"],
  [/:think:/g,      "🤔"],
  [/:skull:/g,      "💀"],
  [/:party:/g,      "🎉"],
];

export function convertEmotes(text: string): string {
  return EMOTES.reduce((s, [pattern, emoji]) => s.replace(pattern, emoji), text);
}

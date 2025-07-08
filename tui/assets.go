package tui

// MonsterArt holds the ASCII art for monsters.
// We can later expand this to hold multiple frames for animation.
var MonsterArt = map[string]string{
	"Гоблин": `
    / \
   (o.o)
   > ~ <
  `,
	"Скелет": `
    .-.
   (o.o)
   |=|
  /   \
 //_ _\\
  `,
	"Огр": `
   ,--.
  (O..O)
 /  ()  \
 \      /
  ` + "`" + `------'
  `,
}


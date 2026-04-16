# STC2STW (Sillytavern Card -> Sillytavern World)

a simple conversion program shamefully vibecoded in golang by OpenAI Codex to turn sillytavern characters into a sillytavern lorebook. oh yeah and it also works for personas probably.

# building

get go. possibly the latest go. i use go from nix, so if you have nix, then `nix-shell -p go`

then clone this repo, and compile it by going:
```bash
go build -o stc2stw cmd/stc2stw/main.go
```

congratz.

# usage

stc2stw by default works by you providing a character card. the most basic usage of stc2stw is:
```bash
./stc2stw <Character Card>.png
```

and the result will be printed in your terminal. You'd typically wanna redirect this to a file somehow... for that you can use `--out` or `-o` like:
```bash
./stc2stw <Character Card>.png -o Some_Creative_Name.json
```

wow. you can now import that right into sillytavern!

you can also put personas from persona backups into a lorebook! on sillytavern you go to Personas > Backup and you get a json...

and if you feed it to stc2stw like thiiis:
```bash
./stc2stw persona_<date>.json --persona (or -p) "user" (case insensitive)
```

stc2stw will make a lorebook with said persona in it.

okay but that's cool. but you're not gonna be expected to manually create a lorebook for e a c h of your characters right?

surprise we have mass mode `--mass`

so you can literally go:
```bash
./stc2stw Character1.png Character2.json (oh yeah btw character json exports are supported too) --mass -o party.json
```

persona syntax changes in mass mode as `--persona` is disabled. instead you will need to define personas like `<Source export file>.json:<Persona name>`

as a better example:
```bash
./stc2stw Character1.png Character2.json persona_<date>.json:Mario (also case insensitive) --mass
```

# compacting embedded lorebooks

by default stc2stw will also compact embedded lorebooks found inside character cards or personas into the generated output lorebook.

that means:
- the normal main entry for the character/persona is still generated
- embedded WI entries are copied over too
- copied entries are renamed like `(src: Alice) -> Capital City`
- the source character/persona name is added as a secondary trigger so the copied lore is scoped to that source

if you want the old behavior and only want the generated main entry, use:
```bash
./stc2stw Character.png --no-compact
```

# configuring insertion position

by default stc2stw will make all entries be inserted as "@D 👤", aka as a user role message. usually this is fine! but sometimes on some models and presets, this might actually be **evil**, and may cause the AI to have no idea what the fuck is going on.

luckily there's a `--position` or `-P`!

by default it's set to "@duser" or "@D 👤", god i hate typing emojis in vim. but here is precisely all the options you have:

- `bchar`: Before Char Defs
- `achar`: After Char Defs
- `bex`: Before Example Messages
- `aex`: After Example Messages
- `tan`: Top of AN
- `ban`: Bottom of AN
- `@dsys`: @D ⚙️
- `@duser`: @D 👤
- `@dass`: @D 🤖
- `outlet`: Outlet (use with {{outlet::stc2stw}} :3)

note: the meaning of these can be found on the sillytavern docs, here: [Insertion Position](https://docs.sillytavern.app/usage/core-concepts/worldinfo/#insertion-position)

now if you ask me on how the fuck do these work. i'll have no clue. lorebooks are still arcane magic for me.

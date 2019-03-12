# Kablature - Convert text files to kalimba tablature images

The program takes a custom text file format and generates an SVG image
describing the song.

## Building and Usage

Usage:

```
$ go get github.com/samuel-hunter/kablature
$ go install github.com/samuel-hunter/kablature
```

To generate a basic song:

```
$ cat > hot-cross-buns.kab <<EOF
2 b a 4 g
2 b a 4 g
1 g g g g a a a a
2 b a 4 g
EOF
$ kablature -i hot-cross-buns.kab -o hot-cross-buns.svg
```

*Note:* `-o hot-cross-buns.svg` is optional. Without an argument, it
 will by default save to `out.svg`.

## Language

```
$ cat > song.kab <<<EOF
# Eighth notes
1 c d e f g h i j
# Quarter notes
2 c d e f
# Half notes
4 c d
# Whole notes
8 c

# Dotted
4. c 2 d
# Octave shifting
2 c > d e < f # D and E are an octave higher
# Raised note
2 c d e' f # the E is an octave higher
# Rests
1 r r 2 r 4 r
# Chords
4 (c e g) (a f c)
```

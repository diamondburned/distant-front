{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation rec {
	name = "distant-server";
	buildInputs = with pkgs; [ deno ];

	shellHook = ''
		PATH="$HOME/.deno/bin:$PATH"
	'';
}

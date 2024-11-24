import Aura from "@primevue/themes/aura";
import { definePreset, palette } from "@primevue/themes";

const primaryColorPalette = palette("#b4bd00");
const secondaryColorPalette = palette("#64748b");

const blueColorPalette = palette("#3b82f6");
const greenColorPalette = palette("#22c55e");
const yellowColorPalette = palette("#eab308");
const cyanColorPalette = palette("#06b6d4");
const pinkColorPalette = palette("#ec4899");
const indigoColorPalette = palette("#6366f1");
const tealColorPalette = palette("#14b8a6");
const orangeColorPalette = palette("#f7a400");
const purpleColorPalette = palette("#a855f7");
const redColorPalette = palette("#ff3d32");

const LiquiswissTheme = definePreset(Aura, {
  // semantic: {
  //   primary: primaryColorPalette,
  // },
  // primitive: {
  //   blue: blueColorPalette,
  //   sky: blueColorPalette,
  //   green: greenColorPalette,
  //   yellow: yellowColorPalette,
  //   cyan: cyanColorPalette,
  //   pink: pinkColorPalette,
  //   indigo: indigoColorPalette,
  //   teal: tealColorPalette,
  //   orange: orangeColorPalette,
  //   purple: purpleColorPalette,
  //   red: redColorPalette,
  //   slate: secondaryColorPalette,
  // },
});

export default LiquiswissTheme;

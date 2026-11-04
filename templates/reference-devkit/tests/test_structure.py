"""Structure of the devkits repository; standard library only (python3 -m unittest discover tests)."""

import json
import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
NAME = re.compile(r"[a-z0-9]([a-z0-9-]{0,40}[a-z0-9])?")


class Structure(unittest.TestCase):
    def test_base_is_pinned_by_digest(self):
        image = json.loads((ROOT / "base.json").read_text())["image"]
        self.assertRegex(image, r".+@sha256:[0-9a-f]{64}", "put the digest of your base image into base.json")

    def test_every_feature_is_complete(self):
        for directory in (ROOT / "features").iterdir():
            with self.subTest(feature=directory.name):
                self.assertRegex(directory.name, NAME)
                self.assertTrue((directory / "install.sh").is_file())
                self.assertEqual(json.loads((directory / "feature.json").read_text())["id"], directory.name)

    def test_every_kit_uses_existing_features(self):
        for kit in (ROOT / "kits").glob("*.json"):
            spec = json.loads(kit.read_text())
            with self.subTest(kit=kit.stem):
                self.assertRegex(kit.stem, NAME)
                self.assertRegex(spec["apt_snapshot"], r"\d{8}T\d{6}Z")
                self.assertTrue(spec["features"])
                self.assertEqual(len(spec["features"]), len(set(spec["features"])))
                for feature in spec["features"]:
                    self.assertTrue((ROOT / "features" / feature / "install.sh").is_file(), feature)


if __name__ == "__main__":
    unittest.main()

# OCI network boot files CLI tool

This tool helps to push and pull files required for network booting (PXE, BOOTP, HTTP-EFI) in order to install operating systems, images or bootable containers to and from OCI-compatible container registries. Artifacts can be signed and distributed in dedicated repositories or with bootable containers.

For a Linux installation, typically these files are:

* Shim bootloader
* A bootloader (e.g. grub2)
* Linux kernel
* Linux initramdisk

This is work in progress.

## Building

No binaries are currently available, you need Go 1.21+ to build the CLI tool:

    git clone https://github.com/osbuild/oci-netboot
    cd oci-netboot
    go build ./cmd/nboci/

## Preparing boot files

Getting boot files is different depending on the target operating system. Here is an example for Fedora installation ISO:

    wget http://fedora/Fedora-netboot-XX.iso
    7z x Fedora-netboot-XX.iso
    sha256sum EFI/BOOT/BOOTX64.EFI EFI/BOOT/grubx64.efi images/pxeboot/vmlinuz images/pxeboot/initrd.img

In the examples below, these files will be pushed:

    2b7918e408b5cce5e9df3018c55e426493d8327d68694ce0bc29802c05decadf  EFI/BOOT/BOOTX64.EFI
    2db12e047966f19323f815da155d7bf39c4f832aa0b2fbe1f5dca38152c106a2  EFI/BOOT/grubx64.efi
    63f6ee372f74353dd90ac5287a621d8a41694369ee6241f2bd980a3537d19789  images/pxeboot/vmlinuz
    f0aed0be4c2f68c97c320717fc223ec6feb2688473a2785f3e09a53ed240d8db  images/pxeboot/initrd.img

The same files can be also found in RPM packages named `shim-x64` and `grub2-efi-x64`. Additionally, PXELinux bootloader for legacy (BIOS) systems can be only found in `syslinux-tftpboot` package.

It is recommended to copy them into single directory and rename `BOOTX64.EFI` to just `shim.efi`.

## Pushing boot files

Each push operation requires the following input:

* Operating system name (lowercase, alphanum)
* Operating system version (lowercase, alphanum, dots allowed)
* Operating system architecture (lowercase, alphanum, underscore allowed)
* Entrypoint: bootoader filename as the entrypoint (e.g. when SecureBoot is on on x86_64)
* Alternate entrypoint: bootoader filename as the alternative entrypoint (e.g. when SecureBoot is off on x86_64)
* Legacy entrypoint: bootoader filename for legacy systems (e.g. BIOS mode on x86_64) 

Example:

    ./nboci push --repository ghcr.io/lzap/bootc-netboot-example \
        --osname rhel \
        --osversion 9.3.0 \
        --osarch x86_64 \
        --entrypoint shim.efi \
        --alt-entrypoint grubx64.efi \
        --legacy-entrypoint pxelinux.0 \
        shim.efi grubx64.efi pxelinux.0 vmlinuz initrd.img

Files are compressed via zstd and pushed, content is tagged with the following tag:

    osname-osversion-osarch

Tag name can ve overriden with `--tag` argument.

Other examples:

    ./nboci --verbose push --repository ghcr.io/lzap/bootc-netboot-example --osname rhel --osversion 9.3.0 --osarch x86_64 --entrypoint shim.efi --alt-entrypoint grubx64.efi fixtures/rhel-9.3.0-x86_64/*

    ./nboci --verbose push --repository ghcr.io/lzap/bootc-netboot-example --osname rhel --osversion 9.3.0 --osarch aarch64 --entrypoint shim.efi --alt-entrypoint grubaa64.efi fixtures/rhel-9.3.0-aarch64/*

## Pulling boot files

To list all tags (including those which are not netboot artifacts):

    ./nboci list ghcr.io/lzap/bootc-netboot-example

To pull all files from all tags:

    ./nboci pull --destination /tmp/test ghcr.io/lzap/bootc-netboot-example

To pull a specific tag use `ghcr.io/lzap/bootc-netboot-example:rhel-9.3.0-x86_64`.

The utility will sychronize files and only download those files which checksums do not match. Entrypoint and alternate entrypoints will be installed as relative symbolink links named `boot` and `boot-alt`.

    tree /tmp/test
    /tmp/test
    └── rhel
        └── 9.3.0
            └── x86_64
                ├── shim.efi
                ├── boot (-> shim.efi)
                ├── grubx64.efi
                ├── boot-alt (-> grubx64.efi)
                ├── pxelinux.0
                ├── boot-legacy (-> pxelinux.0)
                ├── initrd.img
                └── vmlinuz

## Signing files

Commits can be digitally signed using [cosign](https://github.com/sigstore/cosign).

```
cosign generate-key-pair
cosign sign --key cosign.key -y ghcr.io/lzap/bootc-netboot-example:rhel-9.3.0-x86_64
```

To verify content before it is pulled, use `-k` or `--signature-key` option.

    ./nboci pull --signature-key cosign.pub --destination /tmp/test ghcr.io/lzap/bootc-netboot-example

If key is incorrect or signature is missing from the repo, the utility does not download the content.

## How files are stored

This is described in the [specification](https://github.com/ipanova/netboot-oci-specs). Here is an example:

```
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "artifactType": "application/vnd.unknown.artifact.v1",
  "config": {
    "mediaType": "application/vnd.oci.empty.v1+json",
    "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
    "size": 2,
    "data": "e30="
  },
  "layers": [
    {
      "mediaType": "application/x-netboot-file+zstd",
      "digest": "sha256:8c4db4474646a08e4a251e2c1055cf5cf2c1c21f159e9a3ba74a381414652ad9",
      "size": 377793,
      "annotations": {
        "org.opencontainers.image.title": "shim.efi",
        "org.pulpproject.netboot.src.digest": "sha256:32e77976ebbc915f77dd7f15d66a52cb177d5a9d2ee1794b173390b67495c047",
        "org.pulpproject.netboot.src.size": "946736"
      }
    },
    {
      "mediaType": "application/x-netboot-file+zstd",
      "digest": "sha256:8526a40f8b5aa92dd35b01c03f2e748912d56b84c8675fe2f21e36a39b8eb388",
      "size": 565332,
      "annotations": {
        "org.opencontainers.image.title": "grubx64.efi",
        "org.pulpproject.netboot.src.digest": "sha256:735284626212a6267c0e90dab2428e8f82c182af17aec567c80838d219d9fa42",
        "org.pulpproject.netboot.src.size": "2532984"
      }
    },
    {
      "mediaType": "application/x-netboot-file+zstd",
      "digest": "sha256:d50baa5d4bf3af0fabebc7871b884be83f368c83dbe0ce1733087615808a6c15",
      "size": 41637,
      "annotations": {
        "org.opencontainers.image.title": "pxelinux.0",
        "org.pulpproject.netboot.src.digest": "sha256:dfcdf626efa753db88de0bf513c4c2e1c4e46cf084371e294e5f6864f16c2e01",
        "org.pulpproject.netboot.src.size": "42555"
      }
    },
    {
      "mediaType": "application/x-netboot-file+zstd",
      "digest": "sha256:e0180821662f2072771ecfbe4a242b3d7782126d06e1fa965f61e1247485943d",
      "size": 13020437,
      "annotations": {
        "org.opencontainers.image.title": "vmlinuz",
        "org.pulpproject.netboot.src.digest": "sha256:0d7a9a3c4804334b23cd43ffc3aedad4620192d9c520e2f466f56b96aeb2a284",
        "org.pulpproject.netboot.src.size": "13335480"
      }
    },
    {
      "mediaType": "application/x-netboot-file+zstd",
      "digest": "sha256:98328929370bdce755daaba59003399e5554a169feaf3cd9734c3f398f2f5b1c",
      "size": 100870099,
      "annotations": {
        "org.opencontainers.image.title": "initrd.img",
        "org.pulpproject.netboot.src.digest": "sha256:4080a4d952d5145625d18b822214982a87ad981c254fcac671ca9ea245da5e3d",
        "org.pulpproject.netboot.src.size": "102417772"
      }
    }
  ],
  "annotations": {
    "org.pulpproject.netboot.entrypoint": "shim.efi",
    "org.pulpproject.netboot.altentrypoint": "grubx64.efi",
    "org.pulpproject.netboot.legacyentrypoint": "pxelinux.0",
    "org.pulpproject.netboot.os.arch": "x86_64",
    "org.pulpproject.netboot.os.name": "rhel",
    "org.pulpproject.netboot.os.version": "9.3.0"
  }
}
```

## LICENSE

Apache License 2.0

## TODO

* rename repo to nboci
* cosign integration

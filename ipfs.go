package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ipfs/kubo/plugin/loader"
	_ "github.com/ipfs/kubo/plugin/loader"
	_ "github.com/ipfs/kubo/plugin/plugins/badgerds"
	_ "github.com/ipfs/kubo/plugin/plugins/flatfs"

	config "github.com/ipfs/kubo/config"

	core "github.com/ipfs/kubo/core"

	coreapi "github.com/ipfs/kubo/core/coreapi"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/node/libp2p"

	fsrepo "github.com/ipfs/kubo/repo/fsrepo"
)

func init_ipfs(stateDir string) (*core.IpfsNode, iface.CoreAPI, error) {
	ctx := context.Background()
	repoPath := filepath.Join(stateDir, "ipfs")
	pluginPath := filepath.Join(repoPath, "plugins")

	plugins, err := loader.NewPluginLoader(pluginPath)
	if err != nil {
		slog.Error("Failed loading plugin loader", "error", err)
		return nil, nil, err
	}

	err = plugins.Initialize()
	if err != nil {
		slog.Error("Failed initializing plugins", "error", err)
		return nil, nil, err
	}

	err = plugins.Inject()
	if err != nil {
		slog.Error("Failed injecting plugins", "error", err)
		return nil, nil, err
	}

	err = os.MkdirAll(repoPath, 0700)
	if err != nil {
		slog.Error("Failed creating IPFS state directory", "target", repoPath, "error", err)
		return nil, nil, err
	}

	cfg, err := config.Init(io.Discard, 2048)
	if err != nil {
		slog.Error("Failed initialising IPFS config", "error", err)
		return nil, nil, err
	}

	if !fsrepo.IsInitialized(repoPath) {
		err := fsrepo.Init(repoPath, cfg)
		if err != nil {
			slog.Error("Failed initialising IPFS repo", "path", repoPath, "error", err)
			return nil, nil, err
		}
	}

	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		slog.Error("Failed opening IPFS repo", "path", repoPath, "error", err)
		return nil, nil, err
	}

	nodeOptions := &core.BuildCfg{
		Online: true,
		Repo:   repo,
		Host:   libp2p.DefaultHostOption,
	}
	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		slog.Error("Failed initialising IPFS node", "error", err)
		return nil, nil, err
	}

	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		slog.Error("Failed initialising IPFS API", "error", err)

		node.Close()
		return nil, nil, err
	}

	return node, api, nil
}

package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/inconshreveable/log15"
	"github.com/jmichiels/tree"
	"github.com/joomcode/errorx"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var _ tree.Node = &Node{}
var _ tree.Tree = &CertificateForest{}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "A short description of tree",
	Long:  `A longer description of tree`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log15.New(log15.Ctx{"module": "tree"})
		// Load the certificates from the cache file
		certs, err := LoadCerts()
		if err != nil {
			logger.Error("Failed to load certificates", "error", err)
			os.Exit(1)
		}
		logger.Info("Loaded certificates", "count", len(certs))
		// Create the certificate forest
		forest := NewCertificateForest(certs)
		logger.Debug("Created certificate forest", "count", len(forest.trees))

		// Print the forest
		tree.Write(forest, os.Stdout)
	},
}

type CertificateForest struct {
	trees     []*Node
	certIndex map[string]*Node
}

func (f *CertificateForest) RootNodes() []tree.Node {
	// typecast into the interface type
	return lo.Map(f.trees, func(node *Node, _ int) tree.Node {
		return tree.Node(node)
	})
}

func (f *CertificateForest) ChildrenNodes(node tree.Node) []tree.Node {
	// typecast into the interface type
	myNode := node.(*Node)
	return lo.Map(myNode.children, func(n *Node, _ int) tree.Node {
		return tree.Node(n)
	})
}

type Node struct {
	self     []CertEntry
	parent   *Node
	children []*Node
}

func NewNode(self CertEntry) *Node {
	return &Node{self: []CertEntry{self}}
}

func (n *Node) AddCertEntry(cert CertEntry) error {
	if len(n.self) == 0 || slices.Equal(n.self[0].Raw, cert.Raw) {
		n.self = append(n.self, cert)
		return nil
	}
	return errorx.IllegalArgument.New("certificate mismatch")
}

func (n *Node) String() string {
	return fmt.Sprintf("%s", n.self[0].Certificate.Subject.String())
}

func LinkNodes(parent, child *Node) {
	if parent == child {
		return
	}
	child.parent = parent
	parent.children = append(parent.children, child)
}

func NewCertificateForest(entries []CertEntry) *CertificateForest {
	forest := &CertificateForest{}
	// certIndex will be used to quickly find certificates by their subject key identifiers
	// the same certificate can be present at multiple places in the filesystem
	certIndex := map[string]*Node{}

	// Create the coalesced nodes. Each certificate corresponds to a node in the forest.
	nodes := lo.Map(entries, func(entry CertEntry, _ int) *Node {
		subjectKeyID := string(entry.Certificate.SubjectKeyId)
		_, ok := certIndex[subjectKeyID]
		if !ok {
			certIndex[subjectKeyID] = NewNode(entry)
		} else {
			certIndex[subjectKeyID].AddCertEntry(entry)
		}
		return certIndex[subjectKeyID]
	})

	// Link the nodes to form a forest
	for _, node := range nodes {
		authKeyId := string(node.self[0].Certificate.AuthorityKeyId)
		if issuer, ok := certIndex[authKeyId]; ok {
			LinkNodes(issuer, node)
		}
	}

	for _, node := range nodes {
		if node.parent == nil || node.parent == node {
			forest.trees = append(forest.trees, node)
		}
	}
	forest.certIndex = certIndex
	return forest
}

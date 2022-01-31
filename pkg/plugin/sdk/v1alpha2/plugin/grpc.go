/*
Copyright © 2021 Aspect Build Systems Inc

Not licensed for re-use.
*/

// grpc.go hides all the complexity of doing the gRPC calls between the aspect
// Core and a Plugin implementation by providing simple abstractions from the
// point of view of Plugin maintainers.
package plugin

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/manifoldco/promptui"
	"google.golang.org/grpc"

	buildeventstream "aspect.build/cli/bazel/buildeventstream/proto"
	"aspect.build/cli/pkg/ioutils"
	"aspect.build/cli/pkg/plugin/sdk/v1alpha2/proto"
)

// GRPCPlugin represents a Plugin that communicates over gRPC.
type GRPCPlugin struct {
	goplugin.Plugin
	Impl Plugin
}

// GRPCServer registers an instance of the GRPCServer in the Plugin binary.
func (p *GRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPluginServer(s, &GRPCServer{Impl: p.Impl, broker: broker})
	return nil
}

// GRPCClient returns a client to perform the RPC calls to the Plugin
// instance from the Core.
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewPluginClient(c), broker: broker}, nil
}

// GRPCServer implements the gRPC server that runs on the Plugin instances.
type GRPCServer struct {
	Impl   Plugin
	broker *goplugin.GRPCBroker
}

// BEPEventCallback translates the gRPC call to the Plugin BEPEventCallback
// implementation.
func (m *GRPCServer) BEPEventCallback(
	ctx context.Context,
	req *proto.BEPEventCallbackReq,
) (*proto.BEPEventCallbackRes, error) {
	return &proto.BEPEventCallbackRes{}, m.Impl.BEPEventCallback(req.Event)
}

// PostBuildHook translates the gRPC call to the Plugin PostBuildHook
// implementation. It starts a prompt runner that is passed to the Plugin
// instance to be able to perform prompt actions to the CLI user.
func (m *GRPCServer) PostBuildHook(
	ctx context.Context,
	req *proto.PostBuildHookReq,
) (*proto.PostBuildHookRes, error) {
	conn, err := m.broker.Dial(req.BrokerId)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := proto.NewPrompterClient(conn)
	prompter := &PrompterGRPCClient{client: client}
	return &proto.PostBuildHookRes{},
		m.Impl.PostBuildHook(req.IsInteractiveMode, prompter)
}

// PostBuildHook translates the gRPC call to the Plugin PostBuildHook
// implementation. It starts a prompt runner that is passed to the Plugin
// instance to be able to perform prompt actions to the CLI user.
func (m *GRPCServer) CustomCommands(
	ctx context.Context,
	req *proto.CustomCommandsReq,
) (*proto.CustomCommandsRes, error) {
	fmt.Println("In the GRPCServer CustomCommands")
	conn, err := m.broker.Dial(req.BrokerId)
	if err != nil {
		// should maybe return nil, nil, err
		return nil, nil
	}
	defer conn.Close()

	client := proto.NewPrompterClient(conn)
	prompter := &PrompterGRPCClient{client: client}

	// need to add result to CustomCommandsRes
	customCommands, err := m.Impl.CustomCommands(req.IsInteractiveMode, prompter)
	fmt.Println(customCommands)

	// var network bytes.Buffer
	// enc := gob.NewEncoder(&network) // Will write to network.
	// // dec := gob.NewDecoder(&network) // Will read from network.

	// err2 := enc.Encode(customCommands[0])
	// if err2 != nil {
	// 	fmt.Println("encode error:", err2)
	// }

	// // HERE ARE YOUR BYTES!!!!
	// fmt.Println("1")
	// fmt.Println(network.Bytes())
	// fmt.Println("1")

	// pb := &proto.CustomCommandsRes{
	// 	Command: []*proto.CustomCommandsRes_command{
	// 		{
	// 			SubVariable1: "string",
	// 			SubVariable2: 1,
	// 		},
	// 	},
	// }
	// fmt.Println("--------------------")
	// customCommands[0].Run(nil)
	// fmt.Println("--------------------")

	test := make([][]byte, 0)
	gob.Register(Command{})
	for _, command := range customCommands {
		var network2 bytes.Buffer
		encoder := gob.NewEncoder(&network2) // Will write to network.
		err3 := encoder.Encode(command)
		if err3 != nil {
			fmt.Println("encode error:", err3)
		}
		fmt.Println("2")
		fmt.Println(network2.Bytes())
		fmt.Println("2")
		test = append(test, network2.Bytes())
	}

	pb := &proto.CustomCommandsRes{
		Command: test,
	}

	fmt.Println(pb)

	// // Decode (receive) the value.
	// var q Command
	// err2 = dec.Decode(&q)
	// if err2 != nil {
	// 	fmt.Println("herererererererererere")
	// 	fmt.Println("herererererererererere")
	// 	fmt.Println("herererererererererere")
	// 	fmt.Println("decode error:", err2)
	// }
	// fmt.Printf("My Use name: " + string(q.Use))

	// fmt.Println("2")
	// fmt.Println("2")
	// fmt.Println("2")
	// fmt.Println("2")
	// fmt.Println("2")
	// fmt.Println("2")
	// fmt.Println("2")
	// q.Run([]string{})

	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	// fmt.Println("3")
	return pb, err
}

// PostTestHook translates the gRPC call to the Plugin PostTestHook
// implementation. It starts a prompt runner that is passed to the Plugin
// instance to be able to perform prompt actions to the CLI user.
func (m *GRPCServer) PostTestHook(
	ctx context.Context,
	req *proto.PostTestHookReq,
) (*proto.PostTestHookRes, error) {
	conn, err := m.broker.Dial(req.BrokerId)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := proto.NewPrompterClient(conn)
	prompter := &PrompterGRPCClient{client: client}
	return &proto.PostTestHookRes{},
		m.Impl.PostTestHook(req.IsInteractiveMode, prompter)
}

// PostRunHook translates the gRPC call to the Plugin PostRunHook
// implementation. It starts a prompt runner that is passed to the Plugin
// instance to be able to perform prompt actions to the CLI user.
func (m *GRPCServer) PostRunHook(
	ctx context.Context,
	req *proto.PostRunHookReq,
) (*proto.PostRunHookRes, error) {
	conn, err := m.broker.Dial(req.BrokerId)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := proto.NewPrompterClient(conn)
	prompter := &PrompterGRPCClient{client: client}
	return &proto.PostRunHookRes{},
		m.Impl.PostRunHook(req.IsInteractiveMode, prompter)
}

// GRPCClient implements the gRPC client that is used by the Core to communicate
// with the Plugin instances.
type GRPCClient struct {
	client proto.PluginClient
	broker *goplugin.GRPCBroker
}

// BEPEventCallback is called from the Core to execute the Plugin
// BEPEventCallback.
func (m *GRPCClient) BEPEventCallback(event *buildeventstream.BuildEvent) error {
	_, err := m.client.BEPEventCallback(context.Background(), &proto.BEPEventCallbackReq{Event: event})
	return err
}

// PostBuildHook is called from the Core to execute the Plugin PostBuildHook. It
// starts the prompt runner server with the provided PromptRunner.
func (m *GRPCClient) PostBuildHook(isInteractiveMode bool, promptRunner ioutils.PromptRunner) error {
	prompterServer := &PrompterGRPCServer{promptRunner: promptRunner}
	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		proto.RegisterPrompterServer(s, prompterServer)
		return s
	}
	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)
	req := &proto.PostBuildHookReq{
		BrokerId:          brokerID,
		IsInteractiveMode: isInteractiveMode,
	}
	_, err := m.client.PostBuildHook(context.Background(), req)
	s.Stop()
	return err
}

// CustomCommands is called from the Core to execute the Plugin CustomCommands. It
// starts the prompt runner server with the provided PromptRunner.
func (m *GRPCClient) CustomCommands(isInteractiveMode bool, promptRunner ioutils.PromptRunner) ([]*Command, error) {
	fmt.Println("In the GRPCClient CustomCommands")
	prompterServer := &PrompterGRPCServer{promptRunner: promptRunner}
	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		proto.RegisterPrompterServer(s, prompterServer)
		return s
	}
	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)
	req := &proto.CustomCommandsReq{
		BrokerId:          brokerID,
		IsInteractiveMode: isInteractiveMode,
	}
	customCommandsPB, err := m.client.CustomCommands(context.Background(), req)
	s.Stop()
	// convert the result from CustomCommands

	// var network bytes.Buffer
	// enc := gob.NewEncoder(&network) // Will write to network.
	// dec := gob.NewDecoder(&network) // Will read from network.

	customCommands := make([]*Command, 0)

	for _, commandBytes := range customCommandsPB.Command {
		network2 := bytes.NewBuffer(commandBytes)
		decoder := gob.NewDecoder(network2) // Will write to network.
		var q Command
		err2 := decoder.Decode(&q)
		if err2 != nil {
			fmt.Println("decode error:", err2)
		}
		fmt.Printf("My Use name: " + q.Use)

		customCommands = append(customCommands, &q)
	}

	return customCommands, err
}

// PostTestHook is called from the Core to execute the Plugin PostTestHook. It
// starts the prompt runner server with the provided PromptRunner.
func (m *GRPCClient) PostTestHook(isInteractiveMode bool, promptRunner ioutils.PromptRunner) error {
	prompterServer := &PrompterGRPCServer{promptRunner: promptRunner}
	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		proto.RegisterPrompterServer(s, prompterServer)
		return s
	}
	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)
	req := &proto.PostTestHookReq{
		BrokerId:          brokerID,
		IsInteractiveMode: isInteractiveMode,
	}
	_, err := m.client.PostTestHook(context.Background(), req)
	s.Stop()
	return err
}

// PostRunHook is called from the Core to execute the Plugin PostRunHook. It
// starts the prompt runner server with the provided PromptRunner.
func (m *GRPCClient) PostRunHook(isInteractiveMode bool, promptRunner ioutils.PromptRunner) error {
	prompterServer := &PrompterGRPCServer{promptRunner: promptRunner}
	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		proto.RegisterPrompterServer(s, prompterServer)
		return s
	}
	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)
	req := &proto.PostRunHookReq{
		BrokerId:          brokerID,
		IsInteractiveMode: isInteractiveMode,
	}
	_, err := m.client.PostRunHook(context.Background(), req)
	s.Stop()
	return err
}

// PrompterGRPCServer implements the gRPC server that runs on the Core and is
// passed to the Plugin to allow prompt actions to the CLI user.
type PrompterGRPCServer struct {
	promptRunner ioutils.PromptRunner
}

// Run translates the gRPC call to perform a prompt Run on the Core.
func (p *PrompterGRPCServer) Run(
	ctx context.Context,
	req *proto.PromptRunReq,
) (*proto.PromptRunRes, error) {
	prompt := promptui.Prompt{
		Label:       req.GetLabel(),
		Default:     req.GetDefault(),
		AllowEdit:   req.GetAllowEdit(),
		Mask:        []rune(req.GetMask())[0],
		HideEntered: req.GetHideEntered(),
		IsConfirm:   req.GetIsConfirm(),
		IsVimMode:   req.GetIsVimMode(),
	}

	result, err := p.promptRunner.Run(prompt)
	res := &proto.PromptRunRes{Result: result}
	if err != nil {
		res.Error = &proto.PromptRunRes_Error{
			Happened: true,
			Message:  err.Error(),
		}
	}

	return res, nil
}

// PrompterGRPCClient implements the gRPC client that is used by the Plugin
// instance to communicate with the Core to request prompt actions from the
// user.
type PrompterGRPCClient struct {
	client proto.PrompterClient
}

// Run is called from the Plugin to request the Core to run the given
// promptui.Prompt.
func (p *PrompterGRPCClient) Run(prompt promptui.Prompt) (string, error) {
	label, isString := prompt.Label.(string)
	if !isString {
		return "", fmt.Errorf("label '%+v' must be a string", prompt.Label)
	}
	req := &proto.PromptRunReq{
		Label:       label,
		Default:     prompt.Default,
		AllowEdit:   prompt.AllowEdit,
		Mask:        string(prompt.Mask),
		HideEntered: prompt.HideEntered,
		IsConfirm:   prompt.IsConfirm,
		IsVimMode:   prompt.IsVimMode,
	}
	res, err := p.client.Run(context.Background(), req)
	if err != nil {
		return "", err
	}
	if res.Error != nil && res.Error.Happened {
		return "", fmt.Errorf(res.Error.Message)
	}
	return res.Result, nil
}

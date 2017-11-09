package template

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/wallix/awless/template/internal/ast"
)

func TestWholeCompilation(t *testing.T) {
	tcases := []struct {
		tpl                  string
		expect               string
		expProcessedFillers  map[string]interface{}
		expResolvedVariables map[string]interface{}
	}{
		{
			tpl: `subnetname = my-subnet
vpcref=@vpc
testsubnet = create subnet cidr={test.cidr} vpc=$vpcref name=$subnetname
update subnet id=$testsubnet public=true
instancecount = {instance.count}
create instance subnet=$testsubnet image=ami-12345 count=$instancecount name='my test instance'`,
			expect: `testsubnet = create subnet cidr=10.0.2.0/24 name=my-subnet vpc=vpc-1234
update subnet id=$testsubnet public=true
create instance count=42 image=ami-12345 name='my test instance' subnet=$testsubnet type=t2.micro`,
			expProcessedFillers:  map[string]interface{}{"instance.type": "t2.micro", "test.cidr": "10.0.2.0/24", "instance.count": 42},
			expResolvedVariables: map[string]interface{}{"subnetname": "my-subnet", "vpcref": "vpc-1234", "instancecount": 42},
		},
		{
			tpl: `
create loadbalancer subnets=[sub-1234, sub-2345,@subalias,@subalias] name=mylb
sub1 = create subnet cidr={test.cidr} vpc=@vpc name=subnet1
sub2 = create subnet cidr=10.0.3.0/24 vpc=@vpc name=subnet2
create loadbalancer subnets=[$sub1, $sub2, sub-3456,{backup-subnet}] name=mylb2
`,
			expect: `create loadbalancer name=mylb subnets=[sub-1234,sub-2345,sub-1111,sub-1111]
sub1 = create subnet cidr=10.0.2.0/24 name=subnet1 vpc=vpc-1234
sub2 = create subnet cidr=10.0.3.0/24 name=subnet2 vpc=vpc-1234
create loadbalancer name=mylb2 subnets=[$sub1,$sub2,sub-3456,sub-0987]`,
			expProcessedFillers:  map[string]interface{}{"test.cidr": "10.0.2.0/24", "backup-subnet": "sub-0987"},
			expResolvedVariables: map[string]interface{}{},
		},
		{
			tpl: `
lb0 = create loadbalancer subnets=[sub-1234, sub-2345,@subalias,@subalias] name=mylb
sub1 = create subnet cidr={test.cidr} vpc=@vpc name=subnet1
sub2 = create subnet cidr=10.0.3.0/24 vpc=@vpc name=subnet2
lb1 = create loadbalancer subnets=[$sub1, $sub2, sub-3456,{backup-subnet}] name=mylb2
`,
			expect: `lb0 = create loadbalancer name=mylb subnets=[sub-1234,sub-2345,sub-1111,sub-1111]
sub1 = create subnet cidr=10.0.2.0/24 name=subnet1 vpc=vpc-1234
sub2 = create subnet cidr=10.0.3.0/24 name=subnet2 vpc=vpc-1234
lb1 = create loadbalancer name=mylb2 subnets=[$sub1,$sub2,sub-3456,sub-0987]`,
			expProcessedFillers:  map[string]interface{}{"test.cidr": "10.0.2.0/24", "backup-subnet": "sub-0987"},
			expResolvedVariables: map[string]interface{}{},
		},
		{
			tpl: `
			a = "mysubnet-1"
b = $a
c = {mysubnet2.hole}
d = [$b,$c,{mysubnet3.hole},mysubnet-4]
create loadbalancer subnets=$d name=lb1
e=$b
secondlb = create loadbalancer subnets=[$e,mysubnet-4,{mysubnet5.hole}] name=lb2
`,
			expect: `create loadbalancer name=lb1 subnets=[mysubnet-1,mysubnet-2,mysubnet-3,mysubnet-4]
secondlb = create loadbalancer name=lb2 subnets=[mysubnet-1,mysubnet-4,mysubnet-5]`,
			expProcessedFillers:  map[string]interface{}{"mysubnet2.hole": "mysubnet-2", "mysubnet3.hole": "mysubnet-3", "mysubnet5.hole": "mysubnet-5"},
			expResolvedVariables: map[string]interface{}{"a": "mysubnet-1", "b": "mysubnet-1", "e": "mysubnet-1", "c": "mysubnet-2", "d": []interface{}{"mysubnet-1", "mysubnet-2", "mysubnet-3", "mysubnet-4"}},
		},
		{
			tpl: `
name = instance-{instance.name}-{version}
name2 = my-test-{hole}
create instance image=ami-1234 name=$name subnet=subnet-{version}
create instance image=ami-1234 name=$name2 subnet=sub1234
`,
			expect: `create instance count=42 image=ami-1234 name=instance-myinstance-10 subnet=subnet-10 type=t2.micro
create instance count=42 image=ami-1234 name=my-test-sub-2345 subnet=sub1234 type=t2.micro`,
			expProcessedFillers:  map[string]interface{}{"instance.name": "myinstance", "version": 10, "instance.type": "t2.micro", "instance.count": 42, "hole": "@sub"},
			expResolvedVariables: map[string]interface{}{"name": "instance-myinstance-10", "name2": "my-test-sub-2345"},
		},
		{
			tpl: `
name = "ins$\ta{nce}-"+{instance.name}+{version}
name2 = {hole}+{hole}+"text-with $Special {char-s"
create instance image=ami-1234 name=$name subnet=subnet-{version}
create instance image=ami-1234 name=$name2 subnet=sub1234
`,
			expect: `create instance count=42 image=ami-1234 name='ins$\ta{nce}-myinstance10' subnet=subnet-10 type=t2.micro
create instance count=42 image=ami-1234 name='sub-2345sub-2345text-with $Special {char-s' subnet=sub1234 type=t2.micro`,
			expProcessedFillers:  map[string]interface{}{"instance.name": "myinstance", "version": 10, "instance.type": "t2.micro", "instance.count": 42, "hole": "@sub"},
			expResolvedVariables: map[string]interface{}{"name": "ins$\\ta{nce}-myinstance10", "name2": "sub-2345sub-2345text-with $Special {char-s"},
		},
		{
			tpl: `
create loadbalancer name=mylb subnets={private.subnets}
`,
			expect:               `create loadbalancer name=mylb subnets=[sub-1234,sub-2345]`,
			expProcessedFillers:  map[string]interface{}{"private.subnets": []interface{}{"sub-1234", "sub-2345"}},
			expResolvedVariables: map[string]interface{}{},
		},
		{
			tpl: `
create loadbalancer name=mylb subnets=subnet-1, subnet-2
`,
			expect:               `create loadbalancer name=mylb subnets=[subnet-1,subnet-2]`,
			expProcessedFillers:  map[string]interface{}{},
			expResolvedVariables: map[string]interface{}{},
		}, //retro-compatibility with old list style, without brackets
	}

	for i, tcase := range tcases {
		env := NewEnv()

		env.AddFillers(map[string]interface{}{
			"instance.type":   "t2.micro",
			"test.cidr":       "10.0.2.0/24",
			"instance.count":  42,
			"unused":          "filler",
			"backup-subnet":   "sub-0987",
			"mysubnet2.hole":  "mysubnet-2",
			"mysubnet3.hole":  "mysubnet-3",
			"mysubnet5.hole":  "mysubnet-5",
			"version":         10,
			"instance.name":   "myinstance",
			"hole":            ast.NewAliasValue("sub"),
			"private.subnets": []interface{}{"sub-1234", "sub-2345"},
		})
		env.AliasFunc = func(e, k, v string) string {
			vals := map[string]string{
				"vpc":      "vpc-1234",
				"subalias": "sub-1111",
				"sub":      "sub-2345",
			}
			return vals[v]
		}
		env.DefLookupFunc = func(in string) (Definition, bool) {
			t, ok := DefsExample[in]
			return t, ok
		}

		inTpl := MustParse(tcase.tpl)

		pass := newMultiPass(NormalCompileMode...)

		compiled, _, err := pass.compile(inTpl, env)
		if err != nil {
			t.Fatalf("%d: %s", i+1, err)
		}

		if got, want := compiled.String(), tcase.expect; got != want {
			t.Fatalf("%d: got\n%s\nwant\n%s", i+1, got, want)
		}

		if got, want := env.GetProcessedFillers(), tcase.expProcessedFillers; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: got %v, want %v", i+1, got, want)
		}

		if got, want := env.ResolvedVariables, tcase.expResolvedVariables; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: got %v, want %v", i+1, got, want)
		}
	}
}

type mockCommand struct{ id string }

func (c *mockCommand) ValidateCommand(map[string]interface{}, []string) []error {
	return []error{errors.New(c.id)}
}
func (c *mockCommand) Run(ctx, params map[string]interface{}) (interface{}, error)    { return nil, nil }
func (c *mockCommand) DryRun(ctx, params map[string]interface{}) (interface{}, error) { return nil, nil }
func (c *mockCommand) ValidateParams(p []string) ([]string, error) {
	switch c.id {
	case "1", "2":
		return []string{c.id}, nil
	case "3":
		return []string{c.id}, errors.New("unexpected")
	}
	panic("wooot")
}

func (c *mockCommand) ConvertParams() ([]string, func(values map[string]interface{}) (map[string]interface{}, error)) {
	return []string{"param1", "param2"},
		func(values map[string]interface{}) (map[string]interface{}, error) {
			_, hasParam1 := values["param1"]
			_, hasParam2 := values["param2"]
			if hasParam1 && hasParam2 {
				return map[string]interface{}{"new": fmt.Sprint(values["param1"], values["param2"])}, nil
			}
			return values, nil
		}
}

func TestCommandsPasses(t *testing.T) {
	cmd1, cmd2, cmd3 := &mockCommand{"1"}, &mockCommand{"2"}, &mockCommand{"3"}
	var count int
	env := NewEnv()
	env.Lookuper = func(...string) interface{} {
		count++
		switch count {
		case 1:
			return cmd1
		case 2:
			return cmd2
		case 3:
			return cmd3
		default:
			panic("whaat")
		}
	}

	t.Run("verify commands exist", func(t *testing.T) {
		tpl := MustParse("create instance\nsub = create subnet\ncreate instance")
		count = 0
		_, _, err := verifyCommandsDefinedPass(tpl, env)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("convert params", func(t *testing.T) {
		tpl := MustParse("create instance\nsub = create subnet param1=anything param2=other\ncreate instance param1=anything")
		count = 0
		compiled, _, err := convertParamsPass(tpl, env)
		if err != nil {
			t.Fatal(err)
		}
		exp := map[string]interface{}{"new": "anythingother"}
		if got, want := compiled.CommandNodesIterator()[1].ToDriverParams(), exp; !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v, want %#v", got, want)
		}
		exp = map[string]interface{}{"param1": "anything"}
		if got, want := compiled.CommandNodesIterator()[2].ToDriverParams(), exp; !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	})

	t.Run("validate commands params", func(t *testing.T) {
		tpl := MustParse("create instance\nsub = create subnet\ncreate instance")
		count = 0
		_, _, err := validateCommandsParamsPass(tpl, env)
		if err == nil {
			t.Fatal("expected err got none")
		}
		if got, want := err.Error(), "unexpected"; !strings.Contains(got, want) {
			t.Fatalf("%s should contain %s", got, want)
		}
	})

	t.Run("normalize missing required params as hole", func(t *testing.T) {
		tpl := MustParse("create instance\nsub = create subnet")
		count = 0
		compiled, _, err := normalizeMissingRequiredParamsAsHolePass(tpl, env)
		if err != nil {
			t.Fatal(err)
		}

		for i, cmd := range compiled.CommandNodesIterator() {
			if got, want := len(cmd.GetHoles()), 1; got != want {
				t.Fatalf("%d. got %d, want %d", i+1, got, want)
			}
			if got, want := cmd.GetHoles()[0], fmt.Sprintf("%s.%d", cmd.Entity, i+1); got != want {
				t.Fatalf("%d. got %s, want %s", i+1, got, want)
			}
		}
	})

	t.Run("validate commands", func(t *testing.T) {
		tpl := MustParse("create instance\nsub = create subnet\ncreate instance")
		count = 0
		_, _, err := validateCommandsPass(tpl, env)
		if err == nil {
			t.Fatal("expected err got none")
		}

		checkContainsAll(t, err.Error(), "123")
	})

	t.Run("inject command", func(t *testing.T) {
		count = 0
		tpl := MustParse("create instance\nsub = create subnet\ncreate instance")

		compiled, _, err := injectCommandsPass(tpl, env)
		if err != nil {
			t.Fatal(err)
		}

		cmds := compiled.CommandNodesIterator()

		sameObject := func(got, want interface{}) bool {
			return reflect.ValueOf(got).Pointer() == reflect.ValueOf(want).Pointer()
		}
		if got, want := cmds[0].Command, cmd1; !sameObject(got, want) {
			t.Fatalf("different object: got %#v, want %#v", got, want)
		}
		if got, want := cmds[1].Command, cmd2; !sameObject(got, want) {
			t.Fatalf("different object: got %#v, want %#v", got, want)
		}
		if got, want := cmds[2].Command, cmd3; !sameObject(got, want) {
			t.Fatalf("different object: got %#v, want %#v", got, want)
		}
	})
}

func TestExternallyProvidedParams(t *testing.T) {
	tcases := []struct {
		template            string
		externalParams      string
		expect              string
		expProcessedFillers map[string]interface{}
	}{
		{
			template:            `create instance count=1 image=ami-123 name=test subnet={hole.name} type=t2.micro`,
			externalParams:      "hole.name=subnet-2345",
			expect:              `create instance count=1 image=ami-123 name=test subnet=subnet-2345 type=t2.micro`,
			expProcessedFillers: map[string]interface{}{"hole.name": "subnet-2345"},
		},
		{
			template:            `create instance count=1 image=ami-123 name=test subnet={hole.name} type={instance.type}`,
			externalParams:      "instance.type=t2.nano hole.name=@subalias",
			expect:              `create instance count=1 image=ami-123 name=test subnet=subnet-111 type=t2.nano`,
			expProcessedFillers: map[string]interface{}{"hole.name": "@subalias", "instance.type": "t2.nano"},
		},
		{
			template:            `create loadbalancer name=elbv2 subnets={my.subnets}`,
			externalParams:      "my.subnets=[@sub1, @sub2]",
			expect:              `create loadbalancer name=elbv2 subnets=[subnet-123,subnet-234]`,
			expProcessedFillers: map[string]interface{}{"my.subnets": []string{"@sub1", "@sub2"}},
		},
		{
			template:            `create loadbalancer name={my.name} subnets={my.subnets}`,
			externalParams:      "my.subnets=sub1, sub2 my.name=loadbalancername",
			expect:              `create loadbalancer name=loadbalancername subnets=[sub1,sub2]`,
			expProcessedFillers: map[string]interface{}{"my.name": "loadbalancername", "my.subnets": []interface{}{"sub1", "sub2"}},
		}, //retro-compatibility with old list style, without brackets
	}
	for i, tcase := range tcases {
		env := NewEnv()
		env.AliasFunc = func(e, k, v string) string {
			vals := map[string]string{
				"subalias": "subnet-111",
				"sub1":     "subnet-123",
				"sub2":     "subnet-234",
			}
			return vals[v]
		}
		env.DefLookupFunc = func(in string) (Definition, bool) {
			t, ok := DefsExample[in]
			return t, ok
		}
		externalFillters, err := ParseParams(tcase.externalParams)
		if err != nil {
			t.Fatal(err)
		}
		env.Fillers = externalFillters
		inTpl := MustParse(tcase.template)

		pass := newMultiPass(NormalCompileMode...)

		compiled, _, err := pass.compile(inTpl, env)
		if err != nil {
			t.Fatalf("%d: %s", i+1, err)
		}

		if got, want := compiled.String(), tcase.expect; got != want {
			t.Fatalf("%d: got\n%s\nwant\n%s", i+1, got, want)
		}

		if got, want := env.GetProcessedFillers(), tcase.expProcessedFillers; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: got %#v, want %#v", i+1, got, want)
		}
	}
}

func TestInlineVariableWithValue(t *testing.T) {
	env := NewEnv()
	tcases := []struct {
		tpl      string
		expError string
		expTpl   string
	}{
		{"ip = 127.0.0.1\ncreate instance ip=$ip", "", "create instance ip=127.0.0.1"},
		{"ip = 1.2.3.4\ncreate instance ip=$ip\ncreate subnet cidr=$ip", "", "create instance ip=1.2.3.4\ncreate subnet cidr=1.2.3.4"},
	}

	for i, tcase := range tcases {
		inTpl := MustParse(tcase.tpl)

		resolvedTpl, _, err := inlineVariableValuePass(inTpl, env)
		if tcase.expError != "" {
			if err == nil {
				t.Fatalf("%d: expected error, got nil", i+1)
			}
			if got, want := err.Error(), tcase.expError; !strings.Contains(got, want) {
				t.Fatalf("%d: got %s, want %s", i+1, got, want)
			}
			continue
		}
		if got, want := resolvedTpl.String(), tcase.expTpl; got != want {
			t.Fatalf("%d: got\n%s\nwant\n%s", i+1, got, want)
		}
	}
}

func TestDefaultEnvWithNilFunc(t *testing.T) {
	text := "create instance name={instance.name} subnet=@mysubnet"
	env := NewEnv()
	tpl := MustParse(text)

	pass := newMultiPass(resolveHolesPass, resolveMissingHolesPass, resolveAliasPass)

	compiled, _, err := pass.compile(tpl, env)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if got, want := compiled.String(), text; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestBailOnUnresolvedAliasOrHoles(t *testing.T) {
	env := NewEnv()
	tcases := []struct {
		tpl         string
		expAliasErr string
		expHolesErr string
	}{
		{tpl: "create subnet\ncreate instance subnet=@mysubnet name={instance.name}\ncreate instance", expAliasErr: "unresolved alias", expHolesErr: "unresolved holes"},
		{tpl: "create subnet\ncreate instance subnet=@mysubnet\ncreate instance", expAliasErr: "unresolved alias: [mysubnet]"},
		{tpl: "create subnet hole=@myhole\ncreate instance subnet=@mysubnet\ncreate instance", expAliasErr: "unresolved alias: [myhole mysubnet]"},
		{tpl: "create subnet name=subnet\nname=@myinstance\ncreate instance name=$myinstance\ncreate instance", expAliasErr: "unresolved alias: [myinstance]"},
		{tpl: "create subnet\ncreate instance name={instance.name}\ncreate instance", expHolesErr: "unresolved holes: [instance.name]"},
		{tpl: "create subnet\ncreate instance name={instance.name}\ncreate instance\ncreate subnet name={subnet.name}", expHolesErr: "unresolved holes: [instance.name subnet.name]"},
		{tpl: "subnetname = {subnet.name} create subnet name=$subnetname\ncreate instance name=instancename\ncreate instance", expHolesErr: "unresolved holes: [subnet.name]"},
		{tpl: "create subnet\ncreate instance name=instancename\ncreate instance\ncreate subnet subnet=name"},
	}

	for i, tcase := range tcases {
		tpl := MustParse(tcase.tpl)
		_, _, err := failOnUnresolvedAliasPass(tpl, env)
		if err == nil && tcase.expAliasErr != "" {
			t.Fatalf("%d: unresolved aliases: got nil error, expect '%s'", i+1, tcase.expAliasErr)
		} else if err != nil && tcase.expAliasErr == "" {
			t.Fatalf("%d: unresolved aliases: got '%s' error, expect nil", i+1, err.Error())
		} else if got, want := err, tcase.expAliasErr; got != nil && want != "" && !strings.Contains(err.Error(), want) {
			t.Fatalf("%d: unresolved aliases: got '%s', want '%s'", i+1, got.Error(), want)
		}

		_, _, err = failOnUnresolvedHolesPass(tpl, env)
		if err == nil && tcase.expHolesErr != "" {
			t.Fatalf("%d: unresolved holes: got nil error, expect '%s'", i+1, tcase.expHolesErr)
		} else if err != nil && tcase.expHolesErr == "" {
			t.Fatalf("%d: unresolved holes: got '%s' error, expect nil", i+1, err.Error())
		} else if got, want := err, tcase.expHolesErr; got != nil && want != "" && !strings.Contains(err.Error(), want) {
			t.Fatalf("%d: unresolved holes: got '%s', want '%s'", i+1, got.Error(), want)
		}
	}
}

func TestCheckInvalidReferencesDeclarationPass(t *testing.T) {
	env := NewEnv()
	tcases := []struct {
		tpl    string
		expErr string
	}{
		{"sub = create subnet\ninst = create instance subnet=$sub\nip = 127.0.0.1\ncreate instance subnet=$inst ip=$ip", ""},
		{"sub = create subnet\ninst = create instance subnet=$sub\ninst = create instance", "'inst' has already been assigned in template"},
		{"sub = create subnet\ninst = create instance subnet=$sub\ncreate instance subnet=$inst_2", "'inst_2' is undefined in template"},
		{"sub = create subnet\ncreate vpc cidr=10.0.0.0/4", ""},
		{"create instance subnet=$sub\nsub = create subnet", "'sub' is undefined in template"},
		{"create instance\nip = 127.0.0.1", ""},
		{"new_inst = create instance autoref=$new_inst\n", "'new_inst' is undefined in template"},
		{"a = $test", "'test' is undefined in template"},
		{"b = [test1,$test2,{test4}]", "'test2' is undefined in template"},
	}

	for i, tcase := range tcases {
		_, _, err := checkInvalidReferenceDeclarationsPass(MustParse(tcase.tpl), env)
		if tcase.expErr == "" && err != nil {
			t.Fatalf("%d: %v", i+1, err)
		}
		if tcase.expErr != "" && (err == nil || !strings.Contains(err.Error(), tcase.expErr)) {
			t.Fatalf("%d: got %v, expected %s", i+1, err, tcase.expErr)
		}
	}
}

func TestResolveAgainstDefinitionsPass(t *testing.T) {
	env := NewEnv()
	env.DefLookupFunc = func(in string) (Definition, bool) {
		t, ok := DefsExample[in]
		return t, ok
	}

	t.Run("Put definition required param in holes", func(t *testing.T) {
		tpl := MustParse(`create instance type=@custom_type count=$inst_num`)

		resolveAgainstDefinitions(tpl, env)

		assertCmdHoles(t, tpl, map[string][]string{
			"subnet": {"instance.subnet"},
			"image":  {"instance.image"},
		})
		assertCmdAliases(t, tpl, map[string][]string{
			"type": {"custom_type"},
		})
		assertCmdRefs(t, tpl, map[string][]string{
			"count": {"inst_num"},
		})
	})

	t.Run("Err on unexisting templ def", func(t *testing.T) {
		tpl := MustParse(`create none type=t2.micro`)

		_, _, err := resolveAgainstDefinitions(tpl, env)
		if err == nil || !strings.Contains(err.Error(), "createnone") {
			t.Fatalf("expected err with message containing 'createnone'")
		}
	})

	t.Run("Err on unexpected param key", func(t *testing.T) {
		tpl := MustParse(`create instance type=t2.micro
	                        create keypair name={key.name} type=wrong`)

		_, _, err := resolveAgainstDefinitions(tpl, env)
		if err == nil || !strings.Contains(err.Error(), "type") {
			t.Fatalf("expected err with message containing 'type'")
		}
	})

	t.Run("Err on unexpected ref key", func(t *testing.T) {
		tpl := MustParse(`create instance type=t2.micro
		create tag stuff=$any`)

		_, _, err := resolveAgainstDefinitions(tpl, env)
		if err == nil || !strings.Contains(err.Error(), "stuff") {
			t.Fatalf("expected err with message containing 'stuff'")
		}
	})

	t.Run("Err on unexpected hole key", func(t *testing.T) {
		tpl := MustParse(`create instance type=t2.micro
		create tag stuff={stuff.any}`)

		_, _, err := resolveAgainstDefinitions(tpl, env)
		if err == nil || !strings.Contains(err.Error(), "stuff") {
			t.Fatalf("expected err with message containing 'stuff'")
		}
	})
}

func TestResolveMissingHolesPass(t *testing.T) {
	tpl := MustParse(`
	ip = {instance.elasticip}
	create instance subnet={instance.subnet} type={instance.type} name={redis.prod} ip=$ip
	create vpc cidr={vpc.cidr}
	create instance name={redis.prod} id={redis.prod} count=3`)

	var count int
	env := NewEnv()
	env.MissingHolesFunc = func(in string) interface{} {
		count++
		switch in {
		case "instance.subnet":
			return "sub-98765"
		case "redis.prod":
			return "redis-124.32.34.54"
		case "vpc.cidr":
			return "10.0.0.0/24"
		case "instance.elasticip":
			return "1.2.3.4"
		default:
			return ""
		}
	}
	env.AddFillers(map[string]interface{}{"instance.type": "t2.micro"})

	pass := newMultiPass(resolveHolesPass, resolveMissingHolesPass)

	tpl, _, err := pass.compile(tpl, env)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := count, 4; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}

	assertVariableValues(t, tpl,
		"1.2.3.4",
	)
	assertCmdParams(t, tpl,
		map[string]interface{}{"type": "t2.micro", "name": "redis-124.32.34.54", "subnet": "sub-98765"},
		map[string]interface{}{"cidr": "10.0.0.0/24"},
		map[string]interface{}{"id": "redis-124.32.34.54", "name": "redis-124.32.34.54", "count": 3},
	)
}

func TestResolveAliasPass(t *testing.T) {
	tpl := MustParse("create instance subnet=@my-subnet ami={instance.ami} count=3")

	env := NewEnv()
	env.AliasFunc = func(e, k, v string) string {
		vals := map[string]string{
			"my-ami":    "ami-12345",
			"my-subnet": "sub-12345",
		}
		return vals[v]
	}
	env.AddFillers(map[string]interface{}{"instance.ami": ast.NewAliasValue("my-ami")})

	pass := newMultiPass(resolveHolesPass, resolveAliasPass)

	tpl, _, err := pass.compile(tpl, env)
	if err != nil {
		t.Fatal(err)
	}

	assertCmdParams(t, tpl, map[string]interface{}{"subnet": "sub-12345", "ami": "ami-12345", "count": 3})
}

func TestResolveHolesPass(t *testing.T) {
	tpl := MustParse("create instance count={instance.count} type={instance.type}")

	env := NewEnv()
	env.AddFillers(map[string]interface{}{
		"instance.count": 3,
		"instance.type":  "t2.micro",
	})

	tpl, _, err := resolveHolesPass(tpl, env)
	if err != nil {
		t.Fatal(err)
	}

	assertCmdHoles(t, tpl, map[string][]string{})
	assertCmdParams(t, tpl, map[string]interface{}{"type": "t2.micro", "count": 3})
}

func TestCmdErr(t *testing.T) {
	tcases := []struct {
		cmd    *ast.CommandNode
		err    interface{}
		ifaces []interface{}
		expErr error
	}{
		{&ast.CommandNode{Action: "create", Entity: "instance"}, nil, nil, nil},
		{&ast.CommandNode{Action: "create", Entity: "instance"}, "my error", nil, errors.New("create instance: my error")},
		{&ast.CommandNode{Action: "create", Entity: "instance"}, errors.New("my error"), nil, errors.New("create instance: my error")},
		{nil, "my error", nil, errors.New("my error")},
		{&ast.CommandNode{Action: "create", Entity: "instance"}, "my error with %s %d", []interface{}{"Donald", 1}, errors.New("create instance: my error with Donald 1")},
	}
	for i, tcase := range tcases {
		if got, want := cmdErr(tcase.cmd, tcase.err, tcase.ifaces...), tcase.expErr; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: got %#v, want %#v", i+1, got, want)
		}
	}
}

type params map[string]interface{}
type holes map[string][]string
type refs map[string][]string
type aliases map[string][]string

func assertVariableValues(t *testing.T, tpl *Template, exp ...interface{}) {
	for i, decl := range tpl.expressionNodesIterator() {
		if vn, ok := decl.(*ast.ValueNode); ok {
			if got, want := vn.Value.Value(), exp[i]; !reflect.DeepEqual(got, want) {
				t.Fatalf("variables value %d: \ngot\n%v\n\nwant\n%v\n", i+1, got, want)
			}
		}

	}
}

func assertCmdParams(t *testing.T, tpl *Template, exp ...params) {
	for i, cmd := range tpl.CommandNodesIterator() {
		if got, want := params(cmd.ToDriverParams()), exp[i]; !reflect.DeepEqual(got, want) {
			t.Fatalf("params: cmd %d: \ngot\n%v\n\nwant\n%v\n", i+1, got, want)
		}
	}
}

func assertCmdHoles(t *testing.T, tpl *Template, exp ...holes) {
	for i, cmd := range tpl.CommandNodesIterator() {
		h := make(map[string][]string)
		for k, p := range cmd.Params {
			if withHoles, ok := p.(ast.WithHoles); ok && len(withHoles.GetHoles()) > 0 {
				h[k] = withHoles.GetHoles()
			}
		}
		if got, want := holes(h), exp[i]; !reflect.DeepEqual(got, want) {
			t.Fatalf("holes: cmd %d: \ngot\n%v\n\nwant\n%v\n", i+1, got, want)
		}
	}
}

func assertCmdRefs(t *testing.T, tpl *Template, exp ...refs) {
	for i, cmd := range tpl.CommandNodesIterator() {
		r := make(map[string][]string)
		for k, p := range cmd.Params {
			if withRefs, ok := p.(ast.WithRefs); ok && len(withRefs.GetRefs()) > 0 {
				r[k] = withRefs.GetRefs()
			}
		}
		if got, want := refs(r), exp[i]; !reflect.DeepEqual(got, want) {
			t.Fatalf("refs: cmd %d: \ngot\n%v\n\nwant\n%v\n", i+1, got, want)
		}
	}
}

func assertCmdAliases(t *testing.T, tpl *Template, exp ...aliases) {
	for i, cmd := range tpl.CommandNodesIterator() {
		r := make(map[string][]string)
		for k, p := range cmd.Params {
			if withAliases, ok := p.(ast.WithAlias); ok && len(withAliases.GetAliases()) > 0 {
				r[k] = withAliases.GetAliases()
			}
		}
		if got, want := aliases(r), exp[i]; !reflect.DeepEqual(got, want) {
			t.Fatalf("refs: cmd %d: \ngot\n%v\n\nwant\n%v\n", i+1, got, want)
		}
	}
}

func checkContainsAll(t *testing.T, s, chars string) {
	for _, e := range chars {
		if !strings.ContainsRune(s, e) {
			t.Fatalf("%s does not contain '%q'", s, e)
		}
	}
}

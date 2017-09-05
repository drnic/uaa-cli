package cmd_test

import (
	. "code.cloudfoundry.org/uaa-cli/cmd"

	"code.cloudfoundry.org/uaa-cli/config"
	"code.cloudfoundry.org/uaa-cli/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

type TestLauncher struct {
	Target string
}

func (tl *TestLauncher) Run(target string) error {
	tl.Target = target
	return nil
}

var _ = Describe("GetImplicitToken", func() {
	var c uaa.Config
	var ctx uaa.UaaContext

	BeforeEach(func() {
		c = uaa.NewConfigWithServerURL(server.URL())
		config.WriteConfig(c)
		ctx = c.GetActiveContext()
	})

	It("launches a browser for the authorize page and gets the callback params", func() {
		launcher := TestLauncher{}
		run := ImplicitTokenCommandRun(launcher.Run, "openid", 8080)
		go run(nil, []string{"shinyclient"})

		httpClient := &http.Client{}
		// UAA sends the user to this redirect_uri after they auth and grant approvals
		httpClient.Get("http://localhost:8080/?access_token=foo")

		Eventually(launcher.Target).Should(Equal(server.URL() + "/oauth/authorize?client_id=shinyclient&redirect_uri=http%3A%2F%2Flocalhost%3A8080&response_type=token&scope=openid"))
		Eventually(GetSavedConfig().GetActiveContext().AccessToken).Should(Equal("foo"))
		Eventually(GetSavedConfig().GetActiveContext().ClientId).Should(Equal("shinyclient"))
		Eventually(GetSavedConfig().GetActiveContext().GrantType).Should(Equal(uaa.GrantType("implicit")))
	})

	It("handles multiple scopes", func() {
		launcher := TestLauncher{}
		run := ImplicitTokenCommandRun(launcher.Run, "openid,user_attributes", 8081)
		go run(nil, []string{"shinyclient"})

		// Callback from UAA
		httpClient := &http.Client{}
		httpClient.Get("http://localhost:8081/?access_token=foo")

		Eventually(launcher.Target).Should(ContainSubstring("/oauth/authorize?client_id=shinyclient&redirect_uri=http%3A%2F%2Flocalhost%3A8081&response_type=token&scope=openid%2Cuser_attributes"))
		Eventually(GetSavedConfig().GetActiveContext().AccessToken).Should(Equal("foo"))
		Eventually(GetSavedConfig().GetActiveContext().ClientId).Should(Equal("shinyclient"))
		Eventually(GetSavedConfig().GetActiveContext().GrantType).Should(Equal(uaa.GrantType("implicit")))
	})
})

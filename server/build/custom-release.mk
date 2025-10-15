# $(DIST_PATH) is the distribution root directory defined in server/Makefile
# This target customizes files under $(DIST_PATH)/client/

# Throw an error if $(DIST_PATH) is not set
ifndef DIST_PATH
$(error DIST_PATH is not set. Please run this Makefile via server/Makefile or set DIST_PATH manually)
endif

customize-assets:
	# Replace strings using variables
	sed -i '' -e '/"about\.notice"/!{ /"about\.copyright"/!s/Mattermost/$(CUSTOM_JP_PLATFORM_NAME)/g; }' $(DIST_PATH)/client/i18n/ja.*.json || true
	sed -i '' -e 's/GitLab/$(CUSTOM_SERVICE_NAME)/g' -e 's/{service}/$(CUSTOM_SERVICE_NAME)/g' -e '/"about\.notice"/!{ /"about\.copyright"/!s/Mattermost/$(CUSTOM_PLATFORM_NAME)/g; }' $(DIST_PATH)/client/i18n/*.json || true
	sed -i '' -e 's/Mattermost/$(CUSTOM_JP_PLATFORM_NAME)/g' $(DIST_PATH)/client/i18n/ja.json || true
	sed -i '' -e 's/{{.Service}}/$(CUSTOM_SERVICE_NAME)/g' -e 's/Mattermost/$(CUSTOM_PLATFORM_NAME)/g' $(DIST_PATH)/client/i18n/*.json || true

	# Remove GitLab icon from login screen
	icon_str='"svg",{width:"[0-9]\+",height:"[0-9]\+",viewBox:"0 0 [0-9]\+ [0-9]\+",fill:"none",xmlns:"http:\/\/www.w3.org\/2000\/svg","aria-label":t({id:"generic_icons.login.gitlab",defaultMessage:"Gitlab Icon"})}'
	file=$$(grep -l $${icon_str} $(DIST_PATH)/client/*.js || true)
	if [ -n "$$file" ]; then \
		sed -i '' "s|$${icon_str}|\"span\",{}|g" "$$file"; \
		sed -i '' "s/external-login-button-label//g" "$$file"; \
	fi

	# Hide Mattermost logo at the top left (before login)
	hfroute_header='o().createElement("div",{className:c()("hfroute-header",{"has-free-banner":r,"has-custom-site-name":b})}'
	file_hfroute_header=$$(grep -l "$${hfroute_header}" $(DIST_PATH)/client/*.js || true)
	if [ -n "$$file_hfroute_header" ]; then \
		hidden_hfroute_header='o().createElement("div",{className:c()("hfroute-header",{"has-free-banner":r,"has-custom-site-name":b}),style:{visibility:"hidden"}}'; \
		sed -i '' "s|$${hfroute_header}|$${hidden_hfroute_header}|g" "$$file_hfroute_header"; \
	fi

	# Hide loading screen icon
	echo ".LoadingAnimation__compass { display: none; }" >> $(DIST_PATH)/client/css/initial_loading_screen.css
